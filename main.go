package main

import (
	"github.com/apptreesoftware/protoc-gen-twirp_dart/generator"
	"github.com/gogo/protobuf/protoc-gen-gogo/plugin"
	"io"
	"io/ioutil"
	"os"
	"strings"

	gogogen "github.com/gogo/protobuf/protoc-gen-gogo/generator"
	"github.com/golang/protobuf/proto"
)

func main() {
	req := readRequest(os.Stdin)
	writeResponse(os.Stdout, generate(req))
}

func readRequest(r io.Reader) *plugin_go.CodeGeneratorRequest {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		panic(err)
	}

	req := new(plugin_go.CodeGeneratorRequest)
	if err = proto.Unmarshal(data, req); err != nil {
		panic(err)
	}

	if len(req.FileToGenerate) == 0 {
		panic(err)
	}

	return req
}

func generate(in *plugin_go.CodeGeneratorRequest) *plugin_go.CodeGeneratorResponse {
	resp := &plugin_go.CodeGeneratorResponse{}

	gen := gogogen.New()
	gen.Request = in
	gen.WrapTypes()
	gen.SetPackageNames()
	gen.BuildTypeNameMap()
	for _, f := range in.GetProtoFile() {
		// skip google/protobuf/timestamp, we don't do any special serialization for jsonpb.
		if *f.Name == "google/protobuf/timestamp.proto" {
			continue
		}
		cf, err := generator.CreateClientAPI(f, gen)
		if err != nil {
			resp.Error = proto.String(err.Error())
			return resp
		}

		resp.File = append(resp.File, cf)
	}

	//resp.File = append(resp.File, generator.RuntimeLibrary())

	return resp
}

func writeResponse(w io.Writer, resp *plugin_go.CodeGeneratorResponse) {
	data, err := proto.Marshal(resp)
	if err != nil {
		panic(err)
	}
	_, err = w.Write(data)
	if err != nil {

	}
}

type Params map[string]string

func getParameters(in *plugin_go.CodeGeneratorRequest) Params {
	params := make(Params)

	if in.Parameter == nil {
		return params
	}

	pairs := strings.Split(*in.Parameter, ",")

	for _, pair := range pairs {
		kv := strings.Split(pair, "=")
		params[kv[0]] = kv[1]
	}

	return params
}
