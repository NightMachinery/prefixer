# Trailing empty records

`--skip-last-empty-record` defaults to true. After splitting input on the input
separator, prefixer removes exactly one final empty record before transformations,
range selection, or location tracking. Interior empty records remain intact.
Empty input produces no records by default.

Use the bare flag or `--skip-last-empty-record=true` to enable this behavior
explicitly. Use `--skip-last-empty-record=false` to retain the final empty record
and restore the previous behavior. Arguments after `--` remain literal records.

For example, input `a\n\n` with prefix `>` produces `>a\n>` by default,
and `>a\n>\n>` with the option disabled. Output still joins records with the
output separator without appending an extra separator.

The last record is tested before trimming, prefix removal, or replacement.
A nonempty final record transformed into an empty string is retained unless
`--skip-empty` is enabled. Negative ranges such as `--process-include=-1` refer
to the last retained input record.
