# GoRxRename
File and content renaming tool using regex-based rules

## Build

```
make
```

## Usage

```
gorxrename -dir=<directory> -rules=<rules_file> [-ignore=<ignore_file>] [-n]
```

For example:
```
gorxrename -rules example_rules.txt -ignore ignore.txt -dir /home/user/myproject
```

### Ignore List Format

Files or directories ignored by regexes.

### Rules List Format

REGEX=>SUBST

For example:
```
rt_([a-zA-Z]+)=>ccl_$1
```

### Dry run mode

Use flag -n.
