import 'dart:async';
import 'package:http/http.dart';
import 'package:requester/requester.dart';
import 'dart:convert';
import '../model/model.twirp.dart';

abstract class Haberdasher {
  Future<Hat> makeHat(Size size);
  Future<Hat> buyHat(Hat hat);
}

class DefaultHaberdasher implements Haberdasher {
  final String hostname;
  Requester _requester;
  final _pathPrefix = "/twirp/config.service.Haberdasher/";

  DefaultHaberdasher(this.hostname, {Requester requester}) {
    if (requester == null) {
      _requester = new Requester(new Client());
    } else {
      _requester = requester;
    }
  }

  Future<Hat> makeHat(Size size) async {
    var url = "${hostname}${_pathPrefix}MakeHat";
    var uri = Uri.parse(url);
    var request = new Request('POST', uri);
    request.headers['Content-Type'] = 'application/json';
    request.body = json.encode(size.toJson());
    var response = await _requester.send(request);
    if (response.statusCode != 200) {
      throw twirpException(response);
    }
    var value = json.decode(response.body);
    return Hat.fromJson(value);
  }

  Future<Hat> buyHat(Hat hat) async {
    var url = "${hostname}${_pathPrefix}BuyHat";
    var uri = Uri.parse(url);
    var request = new Request('POST', uri);
    request.headers['Content-Type'] = 'application/json';
    request.body = json.encode(hat.toJson());
    var response = await _requester.send(request);
    if (response.statusCode != 200) {
      throw twirpException(response);
    }
    var value = json.decode(response.body);
    return Hat.fromJson(value);
  }

  Exception twirpException(Response response) {
    try {
      var value = json.decode(response.body);
      return new Exception(value['msg']);
    } catch (e) {
      return new Exception(response.body);
    }
  }
}
