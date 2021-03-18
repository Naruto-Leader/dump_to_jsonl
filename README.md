# dump_to_jsonl

Generate JSONL from a `mysqldump` dump file.

## Installation

```
% go get github.com/abetomo/dump_to_jsonl
```

## Usage

```
Usage of dump_to_jsonl:
  -file string
        dump file
  -outdir string
        output directory
```

## Examples

### one_create

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

### two_create

```
% dump_to_jsonl -file ./test/fixtures/two_create.sql
{"id":1,"json":"{\"key\": \"value\"}"}
{"id":2,"json":"{\"no\": 1}"}
{"category_id":1,"created_at":"2020-09-09 10:02:35","description":"description1,'A':\"A\"","id":1,"name":"name1","rate":1.1}
{"category_id":2,"created_at":"2020-09-09 10:02:46","description":"description2,'B':\"B\"","id":2,"name":"name2","rate":2.2}
```

```
% dump_to_jsonl -file ./test/fixtures/two_create.sql -outdir /tmp/two_create

% ls /tmp/two_create
json_table.jsonl  test_table.jsonl

% tail /tmp/two_create/*.jsonl
==> /tmp/two_create/json_table.jsonl <==
{"id":1,"json":"{\"key\": \"value\"}"}
{"id":2,"json":"{\"no\": 1}"}

==> /tmp/two_create/test_table.jsonl <==
{"category_id":1,"created_at":"2020-09-09 10:02:35","description":"description1,'A':\"A\"","id":1,"name":"name1","rate":1.1}
{"category_id":2,"created_at":"2020-09-09 10:02:46","description":"description2,'B':\"B\"","id":2,"name":"name2","rate":2.2}
```
