package generator

// CoreFile is the twirp_dart_core.dart file that gets written
// next to the target file.
var CoreFile = `class TwirpException implements Exception {
  final String message;

  TwirpException(this.message);

  @override
  String toString() {
    return 'TwirpException{message: $message}';
  }
}

class TwirpJsonException extends TwirpException {
  final String code;
  final String msg;
  final dynamic meta;

  TwirpJsonException(this.code, this.msg, this.meta) : super(msg);

  factory TwirpJsonException.fromJson(Map<String, dynamic> json) {
    return TwirpJsonException(
        json['code'] as String, json['msg'] as String, json['meta']);
  }

  @override
  String toString() {
    return 'TwirpJsonException{code: $code, msg: $msg, meta: $meta}';
  }
}
`
