# Gowiki

Gowiki is a single-file single-executable wiki which runs its own webserver.

Writing it was an experiment in determining the cowpaths of small but non-trivial web application development using
[gorilla](http://gorillatoolkit.org), [sqlx](http://github.com/jmoiron/sqlx), and
[modl](http://github.com/jmoiron/modl).

## Features

Gowiki is a simple markdown wiki with the following features:

* markdown syntax w/ github style codeblocks
* distributable as a single binary which runs its own http server
* online customizable styles and page markup
* users and online user signup
* open or closed wikis, user page ownership
* ... by default it looks like [my website](http://jmoiron.net)

Gowiki does not:

* attempt to prevent multiple simultaneous edits
* auto-link WikiWeb style PageNames or require WikiCase urls

TODO:

* keep a revision history
* keep track of page references or interlinking
* add export options mentioned in `/config/files`

## Installing

```
go get github.com/jmoiron/gowiki
```

## Deployment

If your `$GOPATH/bin` is in your `PATH`, you can simply:

```
$ gowiki
Running in deployment mode with bundled resources.
Listening on :2222
```

`GOWIKI_PORT` controls the port to run on and `GOWIKI_PATH` controls the destination of the sqlite database.

Front this with nginx or whatever reverse proxy you like.

### Updating

Gowiki will create static files for things like style sheets and page templates in its database which
you can then edit online to modify the look and behavior of the wiki.  This means that old wikis
which run newer versions of gowiki might not look right.  Gowiki will automatically re-seed these if
they aren't there, so to refresh them, clear out the cache:

```sql
$ sqlite3 wiki.db
...
sqlite> delete from file;
sqlite>
```

## Development

```sh
$ ./gowiki -h
Usage of ./gowiki:
  -db="./wiki.db": path for wiki db
  -debug=false: run with debug mode
  -del-static=false: delete db-cached static files
  -load-static=false: reload db-cached static files
  -port="2222": port to run on
```

If no options are chosen, static files are loaded as necessary and the wiki is run in deploy mode.

