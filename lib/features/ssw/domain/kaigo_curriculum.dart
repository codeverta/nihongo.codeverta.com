class KaigoCurriculum {
  KaigoCurriculum.fromJson(Map<String, dynamic> json)
      : modules = (json['modules'] as List<dynamic>)
            .map((item) => KaigoModule.fromJson(item as Map<String, dynamic>))
            .toList(),
        sources = (json['sources'] as List<dynamic>)
            .map((item) => KaigoSource.fromJson(item as Map<String, dynamic>))
            .toList();

  final List<KaigoModule> modules;
  final List<KaigoSource> sources;
  List<KaigoLesson> get lessons => modules.expand((m) => m.lessons).toList();
  int get minutes => lessons.fold(0, (sum, lesson) => sum + lesson.minutes);
  KaigoLesson? lessonById(String id) {
    for (final lesson in lessons) {
      if (lesson.id == id) return lesson;
    }
    return null;
  }
}

class KaigoModule {
  KaigoModule.fromJson(Map<String, dynamic> json)
      : id = json['id'] as String,
        title = json['title'] as String,
        description = json['description'] as String,
        lessons = (json['lessons'] as List<dynamic>)
            .map((item) => KaigoLesson.fromJson(item as Map<String, dynamic>))
            .toList();

  final String id;
  final String title;
  final String description;
  final List<KaigoLesson> lessons;
}

class KaigoLesson {
  KaigoLesson.fromJson(Map<String, dynamic> json)
      : id = json['id'] as String,
        title = json['title'] as String,
        minutes = json['minutes'] as int,
        objectives = List<String>.from(json['objectives'] as List),
        sections = (json['sections'] as List<dynamic>)
            .map((item) => KaigoSection.fromJson(item as Map<String, dynamic>))
            .toList(),
        steps = List<String>.from(json['steps'] as List),
        caution = json['caution'] as String,
        scenario = json['scenario'] as String,
        response = json['response'] as String,
        vocabulary = (json['vocabulary'] as List<dynamic>)
            .map((item) => KaigoWord.fromJson(item as Map<String, dynamic>))
            .toList(),
        dialogue = (json['dialogue'] as List<dynamic>)
            .map((item) => KaigoPhrase.fromJson(item as Map<String, dynamic>))
            .toList(),
        checks = (json['checks'] as List<dynamic>)
            .map((item) => KaigoCheck.fromJson(item as Map<String, dynamic>))
            .toList();

  final String id;
  final String title;
  final int minutes;
  final List<String> objectives;
  final List<KaigoSection> sections;
  final List<String> steps;
  final String caution;
  final String scenario;
  final String response;
  final List<KaigoWord> vocabulary;
  final List<KaigoPhrase> dialogue;
  final List<KaigoCheck> checks;
}

class KaigoSection {
  KaigoSection.fromJson(Map<String, dynamic> json)
      : title = json['title'] as String,
        body = json['body'] as String;
  final String title;
  final String body;
}

class KaigoWord {
  KaigoWord.fromJson(Map<String, dynamic> json)
      : japanese = json['japanese'] as String,
        reading = json['reading'] as String,
        meaning = json['meaning'] as String;
  final String japanese;
  final String reading;
  final String meaning;
}

class KaigoPhrase extends KaigoWord {
  KaigoPhrase.fromJson(super.json) : super.fromJson();
}

class KaigoCheck {
  KaigoCheck.fromJson(Map<String, dynamic> json)
      : prompt = json['prompt'] as String,
        options = List<String>.from(json['options'] as List),
        answer = json['answer'] as int,
        explanation = json['explanation'] as String;
  final String prompt;
  final List<String> options;
  final int answer;
  final String explanation;
}

class KaigoSource {
  KaigoSource.fromJson(Map<String, dynamic> json)
      : title = json['title'] as String,
        url = json['url'] as String;
  final String title;
  final String url;
}
