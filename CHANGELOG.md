# Changelog

## [2.9.0](https://github.com/stupside/moley/compare/v2.8.0...v2.9.0) (2026-10-28)


### Features

* dress the Go gopher up as a mole for the logo ([57857ea](https://github.com/stupside/moley/commit/57857eacc483c8f25a52098e26f90233085541f6))
* theme the docs after the gopher logo ([f2f1d1d](https://github.com/stupside/moley/commit/f2f1d1d2c6aa7d480477c8e562e8038caffbfa32))

## [2.8.0](https://github.com/stupside/moley/compare/v2.7.1...v2.8.0) (2026-10-27)


### Features

* fall back to the cloudflared login token ([820e9b1](https://github.com/stupside/moley/commit/820e9b167fe40553745db395bc0e84f50186f24a))
* per-app cloudflared origin_request options ([98793d6](https://github.com/stupside/moley/commit/98793d6e218c88f58824e12c3c6f5a0ef7623853))
* route apps by path prefix with expose.path ([37cee73](https://github.com/stupside/moley/commit/37cee73aabbdff55241880624f8418acf3463c27))
* validate Cloudflare Access tokens at the origin ([4e8bb24](https://github.com/stupside/moley/commit/4e8bb24913adf0ff5f23f5de4d55863e6f4cac19))


### Bug Fixes

* bump Go to 1.27.3 and cloudflared to 2026.10.5 ([2d49f8d](https://github.com/stupside/moley/commit/2d49f8d0be6817828502b09902f1e52e40b7d9d0))
* don't warn about the origin check while stopping ([12e2f7e](https://github.com/stupside/moley/commit/12e2f7e7f23cf9b2745770ea3a4d5c5b8cbf4de8))
* migrate to cloudflare-go v6 and paginate Access policy lookup ([598e43a](https://github.com/stupside/moley/commit/598e43a08b44e919a85abc0e8e34ba18952e0ae2))
* name the Access permission origin validation needs ([9732bda](https://github.com/stupside/moley/commit/9732bdafd814a449a8c68eb0a9ea803827ae1b74))
* write credentials-file key that cloudflared actually reads ([696c0c2](https://github.com/stupside/moley/commit/696c0c2a8fdd1b79c95b0b44767b6c646a8c5e45))

## [2.7.1](https://github.com/stupside/moley/compare/v2.7.0...v2.7.1) (2026-10-18)


### Bug Fixes

* upload release assets before publishing immutable releases ([603d77e](https://github.com/stupside/moley/commit/603d77e56df1adc59584d7d5ea55a1774a76cd85))

## [2.7.0](https://github.com/stupside/moley/compare/v2.6.1...v2.7.0) (2026-10-18)


### Features

* update goreleaser setup to latest ([ee8b20b](https://github.com/stupside/moley/commit/ee8b20b3d12f0749a65109446a7092e3a1e2c842))


### Bug Fixes

* allow Windows lock state reads while holding the guard ([5a18bee](https://github.com/stupside/moley/commit/5a18bee13f30252099708532d4d732abd25af275))


### Continuous Integration

* automate releases with release-please ([a10bd90](https://github.com/stupside/moley/commit/a10bd909431033650a7368bc73df081cafd14773))
* keep Dependabot on compatible TypeScript releases ([efcad79](https://github.com/stupside/moley/commit/efcad79b5316f5629a521542401fbab72dd6bba7))
