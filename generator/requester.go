package generator

// RequesterFile is the twirp_dart_core.dart file that gets written
// next to the target file.
var RequesterFile = `import 'dart:async';
import 'package:http/http.dart';

/// Sends [BaseRequest] and returns the [Response] after applying all middleware
class Requester {
  final Client _client;
  final List<Middleware> _middleware;

  Requester(this._client) : _middleware = [];

  Future<Response> send(BaseRequest request) async {
    for (var i in _middleware) {
      i.prepare(request);
    }
    var stream = await _client.send(request);
    var response = await Response.fromStream(stream);

    // Reverse the middleware so that the last middlware to prepare the request
    // is the first to handle the response.
    for (var m in _middleware.reversed) {
      m.handle(response);
    }
    return response;
  }

  void addMiddleware(Middleware m) {
    _middleware.add(m);
  }

  void addAllMiddleware(Iterable<Middleware> m) {
    _middleware.addAll(m);
  }
}

abstract class Middleware {
  void prepare(BaseRequest request);
  void handle(Response response);
}

class BaseMiddleware implements Middleware {
  @override
  void prepare(BaseRequest request) {}
  @override
  void handle(Response response) {}
}
`
