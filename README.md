# crononag

**crononag** is a **super-simple wrapper for commands running from cron** to
limit nagging emails.

## Motivation

There are a couple of existing cron wrappers around, but I needed to be able to
customize the suppressed exit codes, and optionally always supress stdout.  I
also wanted a simple deploy story.  Existing solutions didn't meet these
criteria, so I wrote this.

## Usage

```
crononag [global options] command [command options] [arguments...]
```

Simply [download](https://github.com/pdf/crononag/releases/latest)
`crononag` and put it on your path, then add `crononag` in front of the
executable in your cron definition and enjoy the silence.  You can remove any
output redirections, eg:

```diff
diff a/cronjob b/cronjob
--- a/cronjob
+++ b/cronjob
@@ -1,3 +1,3 @@
 PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

-*/10 * * * * root noisycronjob --with args > /dev/null
+*/10 * * * * root crononag noisycronjob --with args
```

Now, your job will only produce output in the event of a non-zero exit code.
Supressed exit codes may set using the option below.

### Options

Currently, `crononag` supports only very simple options:

```
-E, --suppress-exit-code [-E option -E option]       suppress output for specified exit code(s), may specify multiple times, defaults to 0
-O, --suppress-stdout                                suppress stdout, even on error
```

## License

This project rocks and uses the MIT-LICENSE
