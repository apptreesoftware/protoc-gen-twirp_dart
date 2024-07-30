package generator

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"strings"
	"text/template"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const apiTemplate = `
{{- range .Imports}}
import '{{.Path}}';
{{- end}}

{{- range .Models}}
{{- if not .Primitive}}
class {{.Name}} {

	{{.Name}}(
	{{range .Fields -}}
		this.{{.Name}},
	{{- end}});

    {{range .Fields -}}
	{{ if eq .Name "ID" }}// ignore: non_constant_identifier_names{{ end }}
    {{.Type}}{{ if .IsOptional }}?{{ end }} {{.Name}};
    {{- end }}
	
	factory {{.Name}}.fromJson(Map<String,dynamic> json) {
		{{- range .Fields -}}
			{{if .IsMap}}
			final {{.Name}}Map = {{ mapLiteral .Type }}{};
			(json['{{.JSONName}}'] as Map<String, dynamic>).forEach((key, val) {
				{{if .MapValueField.IsMessage}}
					{{ if isNumber .MapKeyField }}
					{{.Name}}Map[int.parse(key)] = {{.MapValueField.Type}}.fromJson(val as Map<String,dynamic>);
					{{ else }}
					{{.Name}}Map[key] = {{.MapValueField.Type}}.fromJson(val as Map<String,dynamic>);
					{{ end }}
				{{else}}
				if (val is String) {
					{{if eq .MapValueField.Type "double"}}
						{{.Name}}Map[key] = double.parse(val);
					{{end}}
					{{if eq .MapValueField.Type "int"}}
						{{.Name}}Map[key] = int.parse(val);
					{{end}}
				} else if (val is num) {
					{{if eq .MapValueField.Type "double"}}
						{{.Name}}Map[key] = val.toDouble();
					{{end}}
					{{if eq .MapValueField.Type "int"}}
						{{.Name}}Map[key] = val.toInt();
					{{end}}
				}
				{{end}}
			});
			{{end}}
		{{end}}

		return {{.Name}}(
		{{- range .Fields -}}
		{{if .IsMap}}
		{{.Name}}Map,
		{{else if and .IsRepeated .IsMessage}}
		json['{{.JSONName}}'] != null
          ? (json['{{.JSONName}}'] as List)
              .map((d) => {{.InternalType}}.fromJson(d))
              .toList()
          : <{{.InternalType}}>[],
		{{else if .IsRepeated }}
			{{ if .IsInt64 }}
			json['{{.JSONName}}'] != null ? (json['{{.JSONName}}'] as List).map((e) => int.parse(e)).toList() : <{{.InternalType}}>[],
			{{ else  }}
			json['{{.JSONName}}'] != null ? (json['{{.JSONName}}'] as List).cast<{{.InternalType}}>() : <{{.InternalType}}>[],
			{{ end }}
		{{else if and (.IsMessage) (eq .Type "DateTime")}}
		{{.Type}}.parse(json['{{.JSONName}}']),
		{{else if and .IsMessage .IsOptional}}
		json['{{.JSONName}}'] == null ? null : {{.Type}}.fromJson(json['{{.JSONName}}']),
		{{else if .IsMessage}}
		{{.Type}}.fromJson(json['{{.JSONName}}']),
		{{else if .IsInt64}}
		int.parse(json['{{.JSONName}}'] as String), 
		{{else if eq .Type "double"}}
		json['{{.JSONName}}'] != null ? double.parse(json['{{.JSONName}}'].toString()) : 0,
		{{else}}
		json['{{.JSONName}}'] as {{.Type}},
		{{- end}}
		{{- end}}
		);
	}

	Map<String,dynamic>toJson() {
		final map = <String, dynamic>{};
    	{{- range .Fields -}}
		{{- if .IsMap }}
		map['{{.JSONName}}'] = json.decode(json.encode({{.Name}}));
		{{- else if and .IsRepeated .IsMessage}}
		map['{{.JSONName}}'] = {{.Name}}.map((l) => l.toJson()).toList();
		{{- else if .IsRepeated }}
		map['{{.JSONName}}'] = {{.Name}}.map((l) => l).toList();
		{{- else if and (.IsMessage) (eq .Type "DateTime")}}
		map['{{.JSONName}}'] = {{.Name}}.toIso8601String();
		{{- else if and .IsMessage .IsOptional}}
		map['{{.JSONName}}'] = {{.Name}}?.toJson();
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
{{end -}}

{{range .Services}}
abstract class {{.Name}} {
	{{- range .Methods}}
	Future<{{.OutputType}}>{{.Name}}({{.InputType}} {{.InputArg}});
    {{- end}}
}

typedef Execute = Future<Response> Function(Request);

typedef Middleware = Execute Function(API api, String methodName, Execute mw);

class Default{{.Name}} implements {{.Name}} {
	final String hostname;
    Requester _requester = Requester(Client());
	final _pathPrefix = "/twirp/{{.Package}}.{{.Name}}/";
	final List<Middleware> middleware;

	DefaultAPI(
		this.hostname, {
		Requester? requester,
		this.middleware = const [],
	}) {
		if (requester != null) {
		_requester = requester;
		}
	}

    {{range .Methods}}
	@override
	Future<{{.OutputType}}>{{.Name}}({{.InputType}} {{.InputArg}}) async {
		final url = "$hostname${_pathPrefix}{{.Path}}";
		final uri = Uri.parse(url);
    	final request = Request('POST', uri);
		request.headers['Content-Type'] = 'application/json';
    	request.body = json.encode({{.InputArg}}.toJson());
		Execute f;
		f = (BaseRequest request) async {
			return await _requester.send(request);
		};
		for (final m in middleware) {
			f = m(this, '{{ .Name }}', f);
		}
		final response = await f(request);
		if (response.statusCode != 200) {
     		throw twirpException(response);
    	}
    	final value = json.decode(response.body);
    	return {{.OutputType}}.fromJson(value);
	}
    {{end}}

	Exception twirpException(Response response) {
    	try {
      		final value = json.decode(response.body);
      		return TwirpJsonException.fromJson(value);
    	} catch (e) {
      		return TwirpException(response.body);
    	}
  	}
}

{{end}}

`

type Model struct {
	Name      string
	Primitive bool
	Fields    []ModelField
}

type ModelField struct {
	Name          string
	Type          string
	InternalType  string
	JSONName      string
	JSONType      string
	IsInt64       bool
	IsMessage     bool
	IsRepeated    bool
	IsMap         bool
	IsOptional    bool
	MapKeyField   *ModelField
	MapValueField *ModelField
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
	Imports     []Import
	modelLookup map[string]*Model
}

type Import struct {
	Path string
}

func (ctx *APIContext) AddModel(m *Model) {
	ctx.Models = append(ctx.Models, m)
	ctx.modelLookup[m.Name] = m
}

func (ctx *APIContext) ApplyImports(f *protogen.File) {
	var deps []Import

	if len(ctx.Services) > 0 {
		deps = append(
			deps,
			Import{"dart:async"},
			Import{"package:http/http.dart"},
			Import{"./requester.dart"},
			Import{"./twirp_dart_core.dart"},
		)
	}
	deps = append(deps, Import{"dart:convert"})

	for _, dep := range f.Proto.Dependency {
		if dep == "google/protobuf/timestamp.proto" {
			continue
		}
		importPath := path.Dir(dep)
		sourceDir := path.Dir(f.Proto.GetName())
		sourceComponents := strings.Split(sourceDir, fmt.Sprintf("%c", os.PathSeparator))
		distanceFromRoot := len(sourceComponents)
		for _, pathComponent := range sourceComponents {
			if strings.HasPrefix(importPath, pathComponent) {
				importPath = strings.TrimPrefix(importPath, pathComponent)
				distanceFromRoot--
			}
		}
		fileName := dartFilename(dep)
		fullPath := fileName
		fullPath = path.Join(importPath, fullPath)
		if distanceFromRoot > 0 {
			for i := 0; i < distanceFromRoot; i++ {
				fullPath = path.Join("..", fullPath)
			}
		}
		deps = append(deps, Import{fullPath})
	}
	ctx.Imports = deps
}

func CreateClientAPI(p *protogen.Plugin, f *protogen.File) error {
	ctx := NewAPIContext()
	for _, m := range f.Messages {
		model := &Model{
			Name: string(m.Desc.Name()),
		}
		for _, f := range m.Fields {
			model.Fields = append(model.Fields, newField(f))
		}
		ctx.AddModel(model)
	}
	for _, s := range f.Services {
		service := &Service{
			Name:    string(s.Desc.Name()),
			Package: f.Proto.GetPackage(),
		}

		for _, m := range s.Methods {
			methodPath := string(m.Desc.Name())
			methodName := strings.ToLower(methodPath[0:1]) + methodPath[1:]
			in := string(m.Input.Desc.Name())
			arg := strings.ToLower(in[0:1]) + in[1:]
			service.Methods = append(service.Methods, ServiceMethod{
				Name:       methodName,
				Path:       methodPath,
				InputArg:   arg,
				InputType:  in,
				OutputType: string(m.Output.Desc.Name()),
			})
		}

		ctx.Services = append(ctx.Services, service)
	}

	ctx.AddModel(&Model{
		Name:      "Date",
		Primitive: true,
	})
	ctx.ApplyImports(f)

	funcMap := template.FuncMap{
		"stringify": stringify,
		"parse":     parse,
		"mapLiteral": func(s string) string {
			return strings.TrimPrefix(s, "Map")
		},
		"isNumber": func(m *ModelField) bool {
			return m.JSONType == "number"
		},
	}

	t, err := template.New("client_api").Funcs(funcMap).Parse(apiTemplate)
	if err != nil {
		return err
	}

	b := bytes.NewBufferString("")
	err = t.Execute(b, ctx)
	if err != nil {
		return err
	}

	ff := p.NewGeneratedFile(twirpFilename(f.Proto.GetName()), "")
	_, err = ff.Write(b.Bytes())

	return err
}

func newField(f *protogen.Field) ModelField {
	jsonName := string(f.Desc.Name())
	name := camelCase(jsonName)
	dartType, internalType, jsonType := protoToDartType(f)
	field := ModelField{
		Name:         name,
		Type:         dartType,
		InternalType: internalType,
		JSONName:     jsonName,
		JSONType:     jsonType,
		IsInt64:      f.Desc.Kind() == protoreflect.Int64Kind,
		IsMap:        f.Desc.IsMap(),
		IsRepeated:   f.Desc.IsList(),
		IsOptional:   f.Desc.HasOptionalKeyword(),
	}
	if field.IsMap {
		mapKeyField := newField(f.Message.Fields[0])
		field.MapKeyField = &mapKeyField
		mapValueField := newField(f.Message.Fields[1])
		field.MapValueField = &mapValueField
		field.Type = fmt.Sprintf("Map<%s,%s>", mapKeyField.Type, mapValueField.Type)
	} else { // not sure we require the else here.
		field.IsMessage = f.Desc.Kind() == protoreflect.MessageKind
	}
	return field
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

// generates the (Type, JSONType) tuple for a ModelField so marshal/unmarshal functions
// will work when converting between TS interfaces and protobuf JSON.
func protoToDartType(f *protogen.Field) (string, string, string) {
	dartType := "String"
	jsonType := "string"
	internalType := "String"

	switch f.Desc.Kind() {
	case protoreflect.DoubleKind:
		dartType = "double"
		jsonType = "number"
		break
	case protoreflect.Fixed32Kind,
		protoreflect.Fixed64Kind,
		protoreflect.Int32Kind,
		protoreflect.Int64Kind:
		dartType = "int"
		jsonType = "number"
	case protoreflect.StringKind:
		dartType = "String"
		jsonType = "string"
	case protoreflect.BoolKind:
		dartType = "bool"
		jsonType = "boolean"
	case protoreflect.MessageKind:
		name := string(f.Message.Desc.Name())

		// Google WKT Timestamp is a special case here:
		//
		// Currently the value will just be left as jsonpb RFC 3339 string.
		// JSON.stringify already handles serializing Date to its RFC 3339 format.
		//
		if f.Message.Desc.FullName() == "google.protobuf.Timestamp" {
			dartType = "DateTime"
			jsonType = "string"
		} else {
			dartType = name
			jsonType = name + "JSON"
		}
	}
	internalType = dartType

	if f.Desc.IsList() {
		dartType = "List<" + dartType + ">"
		jsonType = jsonType + "[]"
	}

	return dartType, internalType, jsonType
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
			return fmt.Sprintf("%s.map((n) => Date(n))", field)
		}

		if f.IsMessage {
			return fmt.Sprintf("%s.map(JSONTo%s)", field, singularType)
		}
	}

	if f.Type == "Date" {
		return fmt.Sprintf("Date(%s)", field)
	}

	if f.IsMessage {
		return fmt.Sprintf("JSONTo%s(%s)", f.Type, field)
	}

	return field
}
