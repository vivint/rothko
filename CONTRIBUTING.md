# Contributing

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

### Vgo

An attempt to use vgo and contribute to its development has put this project
in a bit of a weird spot. While building a large part of the project is
possible with vgo, tools such as godocdown and gopherjs don't yet know how to
work without a GOPATH. Thus, in order for these tools to work, you must create
a GOPATH with both the github.com/vivint/rothko and
github.com/gopherjs/gopherjs repos installed. For example, running

```
mkdir rothko
cd rothko
export GOPATH=\`pwd\`
go get github.com/vivint/rothko github.com/gopherjs/gopherjs
cd src/github.com/vivint/rothko
vgo vendor
rm -rf vendor/github.com/gopherjs/gopherjs
```

will get you into an appropriate state.

Sorry for the inconvenience.

## In Elm

- Check out ui/README.md for details on how that's set up.
- Make sure to run elm-format on everything.

## Misc

- If you write bash scripts, make sure they pass shellcheck.

## Finally

Be sure to add yourself to AUTHORS so that you can get credit for your hard
work!
