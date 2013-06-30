# Gowiki

Gowiki is a single-file single-executable wiki which runs its own webserver.

Writing it was an experiment in determining the cowpaths of small but non-trivial web application development using
[gorilla](http://gorillatoolkit.com), [sqlx](http://github.com/jmoiron/sqlx), and
[modl](http://github.com/jmoiron/modl).

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

## Development

Gowiki can run in `DEVELOP` mode by switching a flag in the source.  This will make it load static resources like
javascript, css, and templates off the filesystem.  To run in `DEPLOY` mode again, run `bundle.sh` to re-bundle the
resource changes and then build with `go build`.

