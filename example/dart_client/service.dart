import 'dart:async';
import 'dart:convert';
import 'package:http/http.dart';
import 'package:requester/requester.dart';
import 'twirp.dart';

class Hat {
  Hat(
    this.size,
    this.color,
    this.name,
    this.createdOn,
    this.rgbColor,
    this.availableSizes,
    this.roles,
  );

  int size;
  String color;
  String name;
  DateTime createdOn;
  Color rgbColor;
  List<Size> availableSizes;
  List<int> roles;

  factory Hat.fromJson(Map<String, dynamic> json) {
    return new Hat(
      json['size'] as int,
      json['color'] as String,
      json['name'] as String,
      DateTime.tryParse(json['created_on']),
      new Color.fromJson(json),
      json['availableSizes'] != null
          ? (json['availableSizes'] as List)
              .map((d) => new Size.fromJson(d))
              .toList()
          : <Size>[],
      json['roles'] != null ? (json['roles'] as List).cast<int>() : <int>[],
    );
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
  Color(
    this.red,
    this.green,
    this.blue,
  );

  int red;
  int green;
  int blue;

  factory Color.fromJson(Map<String, dynamic> json) {
    return new Color(
      json['red'] as int,
      json['green'] as int,
      json['blue'] as int,
    );
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
  Receipt(
    this.total,
  );

  double total;

  factory Receipt.fromJson(Map<String, dynamic> json) {
    return new Receipt(
      json['total'] as double,
    );
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
  Size(
    this.inches,
  );

  int inches;

  factory Size.fromJson(Map<String, dynamic> json) {
    return new Size(
      json['inches'] as int,
    );
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
