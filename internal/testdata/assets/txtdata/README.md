# txtdata edge-case fixtures

Hand-written byte-exact TSV samples for the record/field terminator cases of
`assets/txtdata/Readme.md` and `Source/data/file.hpp`. They are small enough to
carry in the tree so the ingest tests never depend on a checkout.

Every file holds the same logical table — header `id<TAB>name`, rows
`1<TAB>alpha` and `2<TAB>beta` — and differs only in encoding.

| Path | Bytes | Case |
|---|---|---|
| `crlf.tsv` | 26 | CRLF records with a trailing CRLF (the upstream corpus encoding) |
| `lf_no_trail.tsv` | 22 | LF records, no trailing LF (`Readme.md`: the final terminator is optional) |
| `bom.tsv` | 26 | UTF-8 BOM followed by LF records (`file.hpp:48-52` strips it) |
| `empty.tsv` | 0 | zero-byte file (`DataFile::Error::NoContent`) |
| `cr.tsv` | 23 | CR-only records — upstream treats a lone CR as a bad record terminator (`data/parser.hpp:115-130`) |

Producer (run once):

```bash
printf 'id\tname\r\n1\talpha\r\n2\tbeta\r\n'      > crlf.tsv
printf 'id\tname\n1\talpha\n2\tbeta'               > lf_no_trail.tsv
printf '\xef\xbb\xbfid\tname\n1\talpha\n2\tbeta\n' > bom.tsv
: > empty.tsv
printf 'id\tname\r1\talpha\r2\tbeta\r'             > cr.tsv
```
