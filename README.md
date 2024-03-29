# kittycad.go

The Golang API client for KittyCAD.

- [Go docs](https://pkg.go.dev/github.com/kittycad/kittycad.go)
- [KittyCAD API Docs](https://zoo.dev/docs/api?lang=go)

## Generating

You can trigger a build with the GitHub action to generate the client. This will
automatically update the client to the latest version based on the spec hosted
at [api.kittycad.io](https://api.kittycad.io/).

Alternatively, if you wish to generate the client locally, run:

```bash
$ make generate
```

## Contributing

Please do not change the code directly since it is generated. PRs that change
the code directly will be automatically closed by a bot.
