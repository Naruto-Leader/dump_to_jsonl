# dump_to_jsonl

Generate JSONL from a `mysqldump` dump file.

## Installation

```
% go get github.com/abetomo/dump_to_jsonl
```

## Usage

```
% dump_to_jsonl -file PATHTO/dump_file.sql
```

## Examples

```
% dump_to_jsonl -file ./test/fixtures/one_create.sql
{"category_id":1,"created_at":"2020-09-09 10:02:35","description":"description1,'A':\"A\"","id":1,"name":"name1","rate":1.1}
{"category_id":2,"created_at":"2020-09-09 10:02:46","description":"description2,'B':\"B\"","id":2,"name":"name2","rate":2.2}
```

```
% cat ./test/fixtures/one_create.sql | dump_to_jsonl
{"category_id":1,"created_at":"2020-09-09 10:02:35","description":"description1,'A':\"A\"","id":1,"name":"name1","rate":1.1}
{"category_id":2,"created_at":"2020-09-09 10:02:46","description":"description2,'B':\"B\"","id":2,"name":"name2","rate":2.2}
```
