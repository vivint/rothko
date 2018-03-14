<img align="right" src="https://raw.githubusercontent.com/vivint/rothko/master/_resources/logo.png">

# Rothko

Rothko stores and allows interaction with distributions of a metric that vary
through time. This allows you to collect insight about the overall values of
metrics when there are many values from multiple hosts.

[Learn more about Rothko in our introductory blog post.](https://innovation.vivint.com/time-series-histograms-with-rothko-observability-for-iot-ce39a09c35fc)

# Contributing

## Set up

Source the `.setup` script to add the `bin` folder to your `$PATH`
(run `. .setup`) which adds the `roth` command. Running `roth` has
output like

```
Usage: roth <subcommand> [subcommand args...]

	build     builds a development rothko binary
	clean     cleans development rothko data
	generate  generates all the code/documentation
	live      runs a live server that rebuilds on code changes
	onboard   sets up the developer environment to build and run the ui
	run       runs a development rothko server
```

Be sure to have [`vgo`](https://research.swtch.com/vgo) installed in your
`$PATH` as well. If you are working on the ui, you can run `roth onboard` to
have it install the required set of npm and Elm dependencies, and `roth live`
to have it spin up a live server (`fswatch` is required). Any changes to the Go
or Elm code will cause the code to be rebuilt and any open web pages to reload.
You can then  visit a demo site with some demo data at
[http://localhost:8080](http://localhost:8080).

## In general

- Pull requests are fine and dandy.
- Try to make the code you add "look like" the code around it. Style and
  consistency matter.
- Open an issue to talk about any major changes you'd like to see. Maybe it's
  already being worked on.

## In Go

- Try to add unit tests for any new functionality you add or any bugs you fix.
  There are some internal packages for doing assertions, etc. Check out other
  tests for guidance.
- Be sure to document any exported public interfaces. Documentation matters.
- Run `roth generate` as it commits documentation to the README's at the
  package directories.
- Breaking changes are still acceptible for now.

## In Elm

- Make sure to run elm-format on everything. I use `elm-format-0.18 0.7.0-exp`.

## Misc

- If you write bash scripts, make sure they pass shellcheck.

## Finally

Be sure to add yourself to `AUTHORS` so that you can get credit for your hard
work!
