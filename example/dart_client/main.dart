import 'dart:async';

import 'service.dart';
import 'twirp.dart';

Future main(List<String> args) async {
  var service = new DefaultHaberdasher('http://localhost:8080');
  try {
    var hat = await service.makeHat(new Size()..inches = 10);
    print(hat);

    hat = await service.makeHat(new Size()..inches = -1);
    print(hat);
  } on TwirpJsonException catch (e) {
    print("${e.code} - ${e.message}");
  } catch (e) {
    print(e);
  }
}
