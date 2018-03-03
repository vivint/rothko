# Rothko

Rothko stores and allows interaction with distributions of a metric that vary
through time. This allows you to collect insight about the overall values of
metrics when there are many values from multiple hosts.

# Installing

If you don't mind getting the latest dependencies, you can run

```
go get github.com/vivint/rothko/bin/rothko
```

If you want to use vgo, you can clone this repo, and run

```
vgo build github.com/vivint/rothko/bin/rothko
```

After you have obtained a binary, it has a number of commands. Run

```
rothko init
```

to have it create a `rothko.toml` in that directory. The file contains some
directives and comments about what they mean. Edit it to your liking, but the
defaults should be good. You can then add some demo data by running

```
rothko demo rothko.toml
```

and then run the server with

```
rothko run rothko.toml
```

which by default listens on localhost:8080.

# Contributing

## Set up

Source the .setup script to add some values to your path (run `. .setup`). Be
sure to have `vgo` installed in your `$PATH`. If you are working on the ui, 
you can run `roth onboard` to have it install the required set of npm 
dependencies, and `roth live` to have it spin up a live server. You can then
visit a demo site with some demo data at localhost:8080.

## In general

- Pull requests are fine and dandy.
- Try to make the code you add "look like" the code around it. Style and
  consistency matter.
- Expect some code review and maybe a bit of back and forth.
- Open an issue to talk about any major changes you'd like to see. Maybe it's
  already being worked on.

## In Go

- Try to add unit tests for any new functionality you add or any bugs you fix.
  There are some internal packages for doing assertions, etc.
- Be sure to document any exported public interfaces. Documentation matters.
- Run ./scripts/docs.sh to update package documentation when you modify any
  exported symbols or their documentation.
- Breaking changes are still acceptible for now.

## In Elm

- Check out ui/README.md for details on how that's set up.
- Make sure to run elm-format on everything.

## Misc

- If you write bash scripts, make sure they pass shellcheck.

## Finally

Be sure to add yourself to AUTHORS so that you can get credit for your hard
work!
