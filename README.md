# Aplos API Client

[![GoDoc](https://pkg.go.dev/badge/github.com/Silicon-Ally/aplos?status.svg)](https://pkg.go.dev/github.com/Silicon-Ally/aplos?tab=doc)
[![CI Workflow](https://github.com/Silicon-Ally/aplos/actions/workflows/test.yml/badge.svg)](https://github.com/Silicon-Ally/aplos/actions?query=branch%3Amain)


Note: This is a pre-v1.0.0 library, expect the API surface to change.

This repo provides a minimal [Aplos API](https://www.aplos.com/api) client in Go, including authentication and a few basic read-only endpoints. Aplos is an online platform for nonprofits + churches to manage their general operations.

The covered API surface is currently quite minimal&mdash;if there's API endpoints or parameters that would be useful to you, feel free to file an issue!

## Usage

```golang

import "github.com/Silicon-Ally/aplos"

...
```

See [the `examples/` directory](/examples) for examples of using the API client.

## Contributing

Contribution guidelines can be found [on our website](https://siliconally.org/oss/contributor-guidelines).
