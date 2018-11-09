import 'dart:async';
import 'dart:convert';
import 'package:http/http.dart';
import 'package:requester/requester.dart';
import 'twirp.dart';

class Hat {
  Hat();
  int size;
  String color;
  String name;
  DateTime createdOn;
  Color rgbColor;
  List<Size> availableSizes;
  List<int> roles;

  factory Hat.fromJson(Map<String, dynamic> json) {
    return new Hat()
      ..size = json['size'] as int
      ..color = json['color'] as String
      ..name = json['name'] as String
      ..createdOn = DateTime.tryParse(json['created_on'])
      ..rgbColor = new Color.fromJson(json)
      ..availableSizes = json['availableSizes'] != null
          ? (json['availableSizes'] as List)
              .map((d) => new Size.fromJson(d))
              .toList()
          : <Size>[]
      ..roles =
          json['roles'] != null ? (json['roles'] as List).cast<int>() : <int>[];
    ;
  }

  Map<String, dynamic> toJson() {
    var map = new Map<String, dynamic>();
    map['size'] = size;
    map['color'] = color;
    map['name'] = name;
    map['created_on'] = createdOn.toIso8601String();
    map['rgbColor'] = rgbColor.toJson();
    map['availableSizes'] = availableSizes?.map((l) => l.toJson())?.toList();

    map['roles'] = roles?.map((l) => l)?.toList();
    return map;
  }

  @override
  String toString() {
    return json.encode(toJson());
  }
}

class Color {
  Color();
  int red;
  int green;
  int blue;

  factory Color.fromJson(Map<String, dynamic> json) {
    return new Color()
      ..red = json['red'] as int
      ..green = json['green'] as int
      ..blue = json['blue'] as int;
  }

  Map<String, dynamic> toJson() {
    var map = new Map<String, dynamic>();
    map['red'] = red;
    map['green'] = green;
    map['blue'] = blue;
    return map;
  }

  @override
  String toString() {
    return json.encode(toJson());
  }
}

class Receipt {
  Receipt();
  double total;

  factory Receipt.fromJson(Map<String, dynamic> json) {
    return new Receipt()..total = json['total'] as double;
  }

  Map<String, dynamic> toJson() {
    var map = new Map<String, dynamic>();
    map['total'] = total;
    return map;
  }

  @override
  String toString() {
    return json.encode(toJson());
  }
}

class Size {
  Size();
  int inches;

  factory Size.fromJson(Map<String, dynamic> json) {
    return new Size()..inches = json['inches'] as int;
  }

  Map<String, dynamic> toJson() {
    var map = new Map<String, dynamic>();
    map['inches'] = inches;
    return map;
  }

  @override
  String toString() {
    return json.encode(toJson());
  }
}

abstract class Haberdasher {
  Future<Hat> makeHat(Size size);
  Future<Hat> buyHat(Hat hat);
}

class DefaultHaberdasher implements Haberdasher {
  final String hostname;
  Requester _requester;
  final _pathPrefix = "/twirp/twitch.twirp.example.Haberdasher/";

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

  TwirpException twirpException(Response response) {
    try {
      var value = json.decode(response.body);
      return new TwirpJsonException.fromJson(value);
    } catch (e) {
      throw new TwirpException(response.body);
    }
  }
}
