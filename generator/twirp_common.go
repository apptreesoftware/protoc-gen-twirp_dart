package generator

import (
	"github.com/golang/protobuf/proto"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
)

func RuntimeLibrary() *plugin.CodeGeneratorResponse_File {
	tmpl := `
class TwirpException {
  final String message;

  TwirpException(this.message);
}

class TwirpJsonException extends TwirpException {
  final String code;
  final String msg;
  final dynamic meta;

  TwirpJsonException(this.code, this.msg, this.meta) : super(msg);

  factory TwirpJsonException.fromJson(Map<String, dynamic> json) {
    return new TwirpJsonException(
        json['code'] as String, json['msg'] as String, json['meta']);
  }
}
`
	cf := &plugin.CodeGeneratorResponse_File{}
	cf.Name = proto.String("twirp.g.dart")
	cf.Content = proto.String(tmpl)

	return cf
}
