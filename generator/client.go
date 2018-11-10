package generator

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"text/template"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
)

const apiTemplate = `
import 'dart:async';
import 'dart:convert';
import 'package:http/http.dart';
import 'package:requester/requester.dart';
import 'twirp.dart';

{{range .Models}}
{{- if not .Primitive}}
class {{.Name}} {

	{{.Name}}(
	{{range .Fields -}}
		this.{{.Name}},
	{{- end}});

    {{range .Fields -}}
    {{.Type}} {{.Name}};
    {{end}}
	
	factory {{.Name}}.fromJson(Map<String,dynamic> json) {
		return new {{.Name}}({{range .Fields -}}
		{{if and .IsRepeated .IsMessage}}
		json['{{.JSONName}}'] != null
          ? (json['{{.JSONName}}'] as List)
              .map((d) => new {{.InternalType}}.fromJson(d))
              .toList()
          : <{{.InternalType}}>[],
		{{else if .IsRepeated }}
		json['{{.JSONName}}'] != null ? (json['{{.JSONName}}'] as List).cast<{{.InternalType}}>() : <{{.InternalType}}>[],
		{{else if and (.IsMessage) (eq .Type "DateTime")}}
		{{.Type}}.tryParse(json['{{.JSONName}}']),
		{{else if .IsMessage}}
		new {{.Type}}.fromJson(json),
		{{else}}
		json['{{.JSONName}}'] as {{.Type}}, 
		{{- end}}
		{{- end}}
		);	
	}

	Map<String,dynamic>toJson() {
		var map = new Map<String, dynamic>();
    	{{- range .Fields -}}
		{{- if and .IsRepeated .IsMessage}}
		map['{{.JSONName}}'] = {{.Name}}?.map((l) => l.toJson())?.toList();
		{{- else if .IsRepeated }}
		map['{{.JSONName}}'] = {{.Name}}?.map((l) => l)?.toList();
		{{- else if and (.IsMessage) (eq .Type "DateTime")}}
		map['{{.JSONName}}'] = {{.Name}}.toIso8601String();
		{{- else if .IsMessage}}
		map['{{.JSONName}}'] = {{.Name}}.toJson();
		{{- else}}
    	map['{{.JSONName}}'] = {{.Name}};
    	{{- end}}
		{{- end}}
		return map;
	}

  @override
  String toString() {
    return json.encode(toJson());
  }
}

{{end -}}
{{end}}

{{range .Services}}
abstract class {{.Name}} {
	{{- range .Methods}}
	Future<{{.OutputType}}>{{.Name}}({{.InputType}} {{.InputArg}});
    {{- end}}
}

class Default{{.Name}} implements {{.Name}} {
	final String hostname;
    Requester _requester;
	final _pathPrefix = "/twirp/{{.Package}}.{{.Name}}/";

    Default{{.Name}}(this.hostname, {Requester requester}) {
		if (requester == null) {
      		_requester = new Requester(new Client());
    	} else {
			_requester = requester;
		}
	}

    {{range .Methods}}
	Future<{{.OutputType}}>{{.Name}}({{.InputType}} {{.InputArg}}) async {
		var url = "${hostname}${_pathPrefix}{{.Path}}";
		var uri = Uri.parse(url);
    	var request = new Request('POST', uri);
		request.headers['Content-Type'] = 'application/json';
    	request.body = json.encode({{.InputArg}}.toJson());
    	var response = await _requester.send(request);
		if (response.statusCode != 200) {
     		throw twirpException(response);
    	}
    	var value = json.decode(response.body);
    	return {{.OutputType}}.fromJson(value);
	}
    {{end}}

  TwirpException twirpException(Response response) {
    try {
      var value = json.decode(response.body);
      return new TwirpJsonException.fromJson(value);
    } catch (e) {
      throw new TwirpException(response.body);
    }
  }
}

{{end}}

`

type Model struct {
	Name         string
	Primitive    bool
	Fields       []ModelField
	CanMarshal   bool
	CanUnmarshal bool
}

type ModelField struct {
	Name         string
	Type         string
	InternalType string
	JSONName     string
	JSONType     string
	IsMessage    bool
	IsRepeated   bool
}

type Service struct {
	Name    string
	Package string
	Methods []ServiceMethod
}

type ServiceMethod struct {
	Name       string
	Path       string
	InputArg   string
	InputType  string
	OutputType string
}

func NewAPIContext() APIContext {
	ctx := APIContext{}
	ctx.modelLookup = make(map[string]*Model)

	return ctx
}

type APIContext struct {
	Models      []*Model
	Services    []*Service
	modelLookup map[string]*Model
}

func (ctx *APIContext) AddModel(m *Model) {
	ctx.Models = append(ctx.Models, m)
	ctx.modelLookup[m.Name] = m
}

// ApplyMarshalFlags will inspect the CanMarshal and CanUnmarshal flags for models where
// the flags are enabled and recursively set the same values on all the models that are field types.
func (ctx *APIContext) ApplyMarshalFlags() {
	for _, m := range ctx.Models {
		for _, f := range m.Fields {
			// skip primitive types and WKT Timestamps
			if !f.IsMessage || f.Type == "DateTime" {
				continue
			}

			baseType := f.Type
			if f.IsRepeated {
				baseType = strings.Replace(baseType, "List<", "", -1)
				baseType = strings.Replace(baseType, ">", "", -1)
			}

			if m.CanMarshal {
				ctx.enableMarshal(ctx.modelLookup[baseType])
			}

			if m.CanUnmarshal {
				m, ok := ctx.modelLookup[baseType]
				if !ok {
					print(baseType)
					log.Fatalf("could not find model of type %s for field %s", baseType, f.Name)
				}
				ctx.enableUnmarshal(m)
			}
		}
	}
}

func (ctx *APIContext) enableMarshal(m *Model) {
	m.CanMarshal = true

	for _, f := range m.Fields {
		// skip primitive types and WKT Timestamps
		if !f.IsMessage || f.Type == "DateTime" {
			continue
		}
		mm, ok := ctx.modelLookup[f.Type]
		if !ok {
			print(f.Name)
			log.Fatalf("could not find model of type %s for field %s", f.Type, f.Name)
		}
		ctx.enableMarshal(mm)
	}
}

func (ctx *APIContext) enableUnmarshal(m *Model) {
	m.CanUnmarshal = true

	for _, f := range m.Fields {
		// skip primitive types and WKT Timestamps
		if !f.IsMessage || f.Type == "DateTime" {
			continue
		}
		mm, ok := ctx.modelLookup[f.Type]
		if !ok {
			print(f.Name)
			log.Fatalf("could not find model of type %s for field %s", f.Type, f.Name)
		}
		ctx.enableUnmarshal(mm)
	}
}

func CreateClientAPI(d *descriptor.FileDescriptorProto) (*plugin.CodeGeneratorResponse_File, error) {
	ctx := NewAPIContext()
	pkg := d.GetPackage()

	// Parse all Messages for generating typescript interfaces
	for _, m := range d.GetMessageType() {
		model := &Model{
			Name: m.GetName(),
		}

		for _, f := range m.GetField() {
			model.Fields = append(model.Fields, newField(f))
		}

		ctx.AddModel(model)
	}

	// Parse all Services for generating typescript method interfaces and default client implementations
	for _, s := range d.GetService() {
		service := &Service{
			Name:    s.GetName(),
			Package: pkg,
		}

		for _, m := range s.GetMethod() {
			methodPath := m.GetName()
			methodName := strings.ToLower(methodPath[0:1]) + methodPath[1:]
			in := removePkg(m.GetInputType())
			arg := strings.ToLower(in[0:1]) + in[1:]

			method := ServiceMethod{
				Name:       methodName,
				Path:       methodPath,
				InputArg:   arg,
				InputType:  in,
				OutputType: removePkg(m.GetOutputType()),
			}

			service.Methods = append(service.Methods, method)
		}

		ctx.Services = append(ctx.Services, service)
	}

	// Only include the custom 'ToJSON' and 'JSONTo' methods in generated code
	// if the Model is part of an rpc method input arg or return type.
	for _, m := range ctx.Models {
		for _, s := range ctx.Services {
			for _, sm := range s.Methods {
				if m.Name == sm.InputType {
					m.CanMarshal = true
				}

				if m.Name == sm.OutputType {
					m.CanUnmarshal = true
				}
			}
		}
	}

	ctx.AddModel(&Model{
		Name:      "Date",
		Primitive: true,
	})

	ctx.ApplyMarshalFlags()

	funcMap := template.FuncMap{
		"stringify": stringify,
		"parse":     parse,
	}

	t, err := template.New("client_api").Funcs(funcMap).Parse(apiTemplate)
	if err != nil {
		return nil, err
	}

	b := bytes.NewBufferString("")
	err = t.Execute(b, ctx)
	if err != nil {
		return nil, err
	}

	cf := &plugin.CodeGeneratorResponse_File{}
	cf.Name = proto.String(tsModuleFilename(d))
	cf.Content = proto.String(b.String())

	return cf, nil
}

func newField(f *descriptor.FieldDescriptorProto) ModelField {
	tsType, internalType, jsonType := protoToTSType(f)
	jsonName := f.GetName()
	name := camelCase(jsonName)

	field := ModelField{
		Name:         name,
		Type:         tsType,
		InternalType: internalType,
		JSONName:     jsonName,
		JSONType:     jsonType,
	}

	field.IsMessage = f.GetType() == descriptor.FieldDescriptorProto_TYPE_MESSAGE
	field.IsRepeated = isRepeated(f)

	return field
}

// generates the (Type, JSONType) tuple for a ModelField so marshal/unmarshal functions
// will work when converting between TS interfaces and protobuf JSON.
func protoToTSType(f *descriptor.FieldDescriptorProto) (string, string, string) {
	dartType := "String"
	jsonType := "string"
	internalType := "String"

	switch f.GetType() {
	case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
		dartType = "double"
		jsonType = "number"
		break
	case descriptor.FieldDescriptorProto_TYPE_FIXED32,
		descriptor.FieldDescriptorProto_TYPE_FIXED64,
		descriptor.FieldDescriptorProto_TYPE_INT32,
		descriptor.FieldDescriptorProto_TYPE_INT64:
		dartType = "int"
		jsonType = "number"
	case descriptor.FieldDescriptorProto_TYPE_STRING:
		dartType = "String"
		jsonType = "string"
	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		dartType = "bool"
		jsonType = "boolean"
	case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		name := f.GetTypeName()

		// Google WKT Timestamp is a special case here:
		//
		// Currently the value will just be left as jsonpb RFC 3339 string.
		// JSON.stringify already handles serializing Date to its RFC 3339 format.
		//
		if name == ".google.protobuf.Timestamp" {
			dartType = "DateTime"
			jsonType = "string"
		} else {
			dartType = removePkg(name)
			jsonType = removePkg(name) + "JSON"
		}
	}
	internalType = dartType

	if isRepeated(f) {
		dartType = "List<" + dartType + ">"
		jsonType = jsonType + "[]"
	}

	return dartType, internalType, jsonType
}

func isRepeated(field *descriptor.FieldDescriptorProto) bool {
	return field.Label != nil && *field.Label == descriptor.FieldDescriptorProto_LABEL_REPEATED
}

func removePkg(s string) string {
	p := strings.Split(s, ".")
	return p[len(p)-1]
}

func camelCase(s string) string {
	parts := strings.Split(s, "_")

	for i, p := range parts {
		if i == 0 {
			parts[i] = p
		} else {
			parts[i] = strings.ToUpper(p[0:1]) + strings.ToLower(p[1:])
		}
	}

	return strings.Join(parts, "")
}

func stringify(f ModelField) string {
	if f.IsRepeated {
		singularType := f.Type[0 : len(f.Type)-2] // strip array brackets from type

		if f.Type == "Date" {
			return fmt.Sprintf("m.%s.map((n) => n.toISOString())", f.Name)
		}

		if f.IsMessage {
			return fmt.Sprintf("m.%s.map(%sToJSON)", f.Name, singularType)
		}
	}

	if f.Type == "Date" {
		return fmt.Sprintf("m.%s.toISOString()", f.Name)
	}

	if f.IsMessage {
		return fmt.Sprintf("%sToJSON(m.%s)", f.Type, f.Name)
	}

	return "m." + f.Name
}

func parse(f ModelField, modelName string) string {
	field := "(((m as " + modelName + ")." + f.Name + ") ? (m as " + modelName + ")." + f.Name + " : (m as " + modelName + "JSON)." + f.JSONName + ")"
	if strings.Compare(f.Name, f.JSONName) == 0 {
		field = "m." + f.Name
	}

	if f.IsRepeated {
		singularType := f.Type[0 : len(f.Type)-2] // strip array brackets from type

		if f.Type == "Date" {
			return fmt.Sprintf("%s.map((n) => new Date(n))", field)
		}

		if f.IsMessage {
			return fmt.Sprintf("%s.map(JSONTo%s)", field, singularType)
		}
	}

	if f.Type == "Date" {
		return fmt.Sprintf("new Date(%s)", field)
	}

	if f.IsMessage {
		return fmt.Sprintf("JSONTo%s(%s)", f.Type, field)
	}

	return field
}
