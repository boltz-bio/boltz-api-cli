# Changelog

## 0.33.0 (2026-06-22)

Full Changelog: [v0.32.0...v0.33.0](https://github.com/boltz-bio/boltz-api-cli/compare/v0.32.0...v0.33.0)

### Features

* **compute-api:** filter list-results by id ([42ba3d1](https://github.com/boltz-bio/boltz-api-cli/commit/42ba3d13c513aef74504aa6754a61bf49f754689))

## 0.32.0 (2026-06-22)

Full Changelog: [v0.31.1...v0.32.0](https://github.com/boltz-bio/boltz-api-cli/compare/v0.31.1...v0.32.0)

### Features

* **auth:** surface organization_name in `auth orgs` ([5bd2660](https://github.com/boltz-bio/boltz-api-cli/commit/5bd26602be5078ee6522f8dac6d26d788c973331))

## 0.31.1 (2026-06-17)

Full Changelog: [v0.31.0...v0.31.1](https://github.com/boltz-bio/boltz-api-cli/compare/v0.31.0...v0.31.1)

### Documentation

* clarify OAuth session cache location ([c1931e7](https://github.com/boltz-bio/boltz-api-cli/commit/c1931e7f8fb1d9c400d6ee18898274190a87bbf6))
* document OS auth storage paths ([3af019a](https://github.com/boltz-bio/boltz-api-cli/commit/3af019a1e05b789a4b04625b66479343f6bbd5da))

## 0.31.0 (2026-06-16)

Full Changelog: [v0.30.0...v0.31.0](https://github.com/boltz-bio/boltz-api-cli/compare/v0.30.0...v0.31.0)

### Features

* **compute-api:** add ADME estimate-cost endpoint ([7ebda2f](https://github.com/boltz-bio/boltz-api-cli/commit/7ebda2f166f05adc1ee20ec64fcf7312147282c9))
* **compute-api:** rename public pipeline identifiers ([7aba459](https://github.com/boltz-bio/boltz-api-cli/commit/7aba45920e769f43f5d638f74158dd8e839de836))


### Bug Fixes

* **boltz2:** route to H100 based on token count, not polymer residues ([004c147](https://github.com/boltz-bio/boltz-api-cli/commit/004c147c0ff84bef7c3d4b24baeb9662519a7434))
* **compute-api:** raise Boltz2 sampling steps minimum ([e77e609](https://github.com/boltz-bio/boltz-api-cli/commit/e77e609390dcb33c84d0e81a4bff1935f3f2da12))
* **compute-api:** remove public model codenames ([3d93517](https://github.com/boltz-bio/boltz-api-cli/commit/3d93517adb512a77d7714273eb36133ab1246083))
* **sm-screen:** batch molecule filtering to avoid 60s Python timeout ([32ced8a](https://github.com/boltz-bio/boltz-api-cli/commit/32ced8abb35a651fa6136fad83609f72a4aa02e1))


### Chores

* **internal:** version bump ([4203ef9](https://github.com/boltz-bio/boltz-api-cli/commit/4203ef97659ed96b825006db01aa4277b6c88f9c))


### Documentation

* **api:** fix runnable predictions example + document request limits ([cc03f02](https://github.com/boltz-bio/boltz-api-cli/commit/cc03f0213e0755e164db9858e107680e2e1c4ea0))
* **api:** use ASCII hyphens in limit descriptions ([d24d4ed](https://github.com/boltz-bio/boltz-api-cli/commit/d24d4edbeacb741f09a233284fd551eb9b5bb4c7))

## 0.30.0 (2026-06-16)

Full Changelog: [v0.29.1...v0.30.0](https://github.com/boltz-bio/boltz-api-cli/compare/v0.29.1...v0.30.0)

### Features

* **compute-api:** cap pending predictions per org and workspace ([2835be7](https://github.com/boltz-bio/boltz-api-cli/commit/2835be7d72d91946ec4aa0d1fde0560118cc6c54))
* **compute:** accept legacy Boltz template inputs ([2519099](https://github.com/boltz-bio/boltz-api-cli/commit/2519099bbbe8626f42fab49548442bae9aa12c36))
* **compute:** canonicalize Boltz template inputs ([c2c03cb](https://github.com/boltz-bio/boltz-api-cli/commit/c2c03cbea7e9cf42a936d1b94c9321f3f4f6ab77))


### Chores

* **internal:** codegen related update ([6c47576](https://github.com/boltz-bio/boltz-api-cli/commit/6c4757694f5480a39f3d3da80dbfc24c1bae3335))

## 0.29.1 (2026-06-16)

Full Changelog: [v0.29.0...v0.29.1](https://github.com/boltz-bio/boltz-api-cli/compare/v0.29.0...v0.29.1)

### Documentation

* **cli:** clean up public README ([de93bbc](https://github.com/boltz-bio/boltz-api-cli/commit/de93bbcad22861313ed63bdd81c0e57d7f4f13a3))

## 0.29.0 (2026-06-15)

Full Changelog: [v0.28.0...v0.29.0](https://github.com/boltz-bio/boltz-api-cli/compare/v0.28.0...v0.29.0)

### Features

* **cli:** add ADME estimate-cost command ([ffdaaf5](https://github.com/boltz-bio/boltz-api-cli/commit/ffdaaf56beadbb57d28beec086019876b2ec8201))


### Bug Fixes

* **cli:** describe adme run input flag ([6e86a97](https://github.com/boltz-bio/boltz-api-cli/commit/6e86a97abc77694d16cf072d7a86a555440b8644))


### Performance Improvements

* **download-results:** stream summary.csv incrementally instead of rewriting per row ([a470e41](https://github.com/boltz-bio/boltz-api-cli/commit/a470e41ba8aec7ebba48d74e8fb5d6d221bb825f))


### Refactors

* **download-results:** drop created_at from summary.csv ([cc9f509](https://github.com/boltz-bio/boltz-api-cli/commit/cc9f509ac6a6ab85661de698f8aa96a50d78c467))
* **download-results:** scope CSV to small-molecule summary ([afcef00](https://github.com/boltz-bio/boltz-api-cli/commit/afcef0098e8565546eb6472775b9ec4cde29e93b))

## 0.28.0 (2026-06-05)

Full Changelog: [v0.27.1...v0.28.0](https://github.com/boltz-bio/boltz-api-cli/compare/v0.27.1...v0.28.0)

### Features

* **lab:** reuse precomputed small molecule pockets ([0edff74](https://github.com/boltz-bio/boltz-api-cli/commit/0edff74423b0ad2154d176f6016c808663d8b1d8))

## 0.27.1 (2026-06-05)

Full Changelog: [v0.27.0...v0.27.1](https://github.com/boltz-bio/boltz-api-cli/compare/v0.27.0...v0.27.1)

### Bug Fixes

* **cli:** describe structure run input flag ([ec519f5](https://github.com/boltz-bio/boltz-api-cli/commit/ec519f52c7a444d9418a9dc06ee40d9e397751c0))

## 0.27.0 (2026-06-05)

Full Changelog: [v0.26.1...v0.27.0](https://github.com/boltz-bio/boltz-api-cli/compare/v0.26.1...v0.27.0)

### Features

* **cli:** add blocking run commands ([be1994f](https://github.com/boltz-bio/boltz-api-cli/commit/be1994f16c0a871efe46b7efc4ac7b2317489cf4))


### Bug Fixes

* **cli:** validate run root before start ([11d7f52](https://github.com/boltz-bio/boltz-api-cli/commit/11d7f520e70da85a8f6651891d589953894d4f32))

## 0.26.1 (2026-06-04)

Full Changelog: [v0.26.0...v0.26.1](https://github.com/boltz-bio/boltz-api-cli/compare/v0.26.0...v0.26.1)

### Bug Fixes

* **download-results:** prefix structures with result IDs ([b3a7b1d](https://github.com/boltz-bio/boltz-api-cli/commit/b3a7b1d1c1b1437c0ff978ea1b5f9f264bf7ad47))

## 0.26.0 (2026-06-04)

Full Changelog: [v0.25.0...v0.26.0](https://github.com/boltz-bio/boltz-api-cli/compare/v0.25.0...v0.26.0)

### Features

* **compute-api:** abort generative pipelines when filter+dedup rate exceeds 95% ([213416e](https://github.com/boltz-bio/boltz-api-cli/commit/213416e97a59d836f3311ff8d035d69678a65be1))

## 0.25.0 (2026-06-03)

Full Changelog: [v0.24.3...v0.25.0](https://github.com/boltz-bio/boltz-api-cli/compare/v0.24.3...v0.25.0)

### Features

* send X-Boltz-Client header for usage attribution ([43dc1df](https://github.com/boltz-bio/boltz-api-cli/commit/43dc1df304f7d069aa68cf7c528848d74364c594))

## 0.24.3 (2026-05-29)

Full Changelog: [v0.24.2...v0.24.3](https://github.com/boltz-bio/boltz-api-cli/compare/v0.24.2...v0.24.3)

### Chores

* **internal:** codegen related update ([6b8ac8d](https://github.com/boltz-bio/boltz-api-cli/commit/6b8ac8d7bc5cdb70980e92244b27a9af1b8ac311))

## 0.24.2 (2026-05-28)

Full Changelog: [v0.24.1...v0.24.2](https://github.com/boltz-bio/boltz-api-cli/compare/v0.24.1...v0.24.2)

### Chores

* **internal:** codegen related update ([0270eab](https://github.com/boltz-bio/boltz-api-cli/commit/0270eab91ee778eab73f2ee69fb3651db8eee571))

## 0.24.1 (2026-05-27)

Full Changelog: [v0.24.0...v0.24.1](https://github.com/boltz-bio/boltz-api-cli/compare/v0.24.0...v0.24.1)

### Documentation

* document structure templates and custom MSA ([187a7db](https://github.com/boltz-bio/boltz-api-cli/commit/187a7db274fafea9a78b2562daea7467df35a191))

## 0.24.0 (2026-05-27)

Full Changelog: [v0.23.0...v0.24.0](https://github.com/boltz-bio/boltz-api-cli/compare/v0.23.0...v0.24.0)

### Features

* [codex] Add custom MSA support to structure predictions ([f977e8c](https://github.com/boltz-bio/boltz-api-cli/commit/f977e8c741d7e9b0200eb07a8a522a2ec28913ad))

## 0.23.0 (2026-05-26)

Full Changelog: [v0.22.1...v0.23.0](https://github.com/boltz-bio/boltz-api-cli/compare/v0.22.1...v0.23.0)

### Features

* Add Boltz2 templates to compute API ([8c84328](https://github.com/boltz-bio/boltz-api-cli/commit/8c843284038fbed9d1be9f4ab2d61c657c16b6fe))

## 0.22.1 (2026-05-26)

Full Changelog: [v0.22.0...v0.22.1](https://github.com/boltz-bio/boltz-api-cli/compare/v0.22.0...v0.22.1)

### Bug Fixes

* **compute-api:** gate pocket conditioning schemas ([3243c7c](https://github.com/boltz-bio/boltz-api-cli/commit/3243c7cb2168380f2205b27370df08fe6f926d1b))

## 0.22.0 (2026-05-26)

Full Changelog: [v0.21.0...v0.22.0](https://github.com/boltz-bio/boltz-api-cli/compare/v0.21.0...v0.22.0)

### Features

* **compute-api:** add protein redesign endpoint ([edb16f6](https://github.com/boltz-bio/boltz-api-cli/commit/edb16f6ae8c2ea51bd095136c48fca148e90148a))

## 0.21.0 (2026-05-25)

Full Changelog: [v0.20.0...v0.21.0](https://github.com/boltz-bio/boltz-api-cli/compare/v0.20.0...v0.21.0)

### Features

* **compute-api:** add small molecule structure templates ([73a7ad4](https://github.com/boltz-bio/boltz-api-cli/commit/73a7ad4f1946d97f2a50ed6e24a4cee87acd7cfc))

## 0.20.0 (2026-05-23)

Full Changelog: [v0.19.1...v0.20.0](https://github.com/boltz-bio/boltz-api-cli/compare/v0.19.1...v0.20.0)

### Features

* **download-results:** parallelize pipeline artifact downloads ([598514f](https://github.com/boltz-bio/boltz-api-cli/commit/598514f2c00e855182074a7abe07e47c5ad319d4))

## 0.19.1 (2026-05-22)

Full Changelog: [v0.19.0...v0.19.1](https://github.com/boltz-bio/boltz-api-cli/compare/v0.19.0...v0.19.1)

### Bug Fixes

* **compute-api:** harden activity retry semantics ([0249c00](https://github.com/boltz-bio/boltz-api-cli/commit/0249c003e7f3d9be1d210b66ad49d539574c322f))

## 0.19.0 (2026-05-18)

Full Changelog: [v0.18.4...v0.19.0](https://github.com/boltz-bio/boltz-api-cli/compare/v0.18.4...v0.19.0)

### Features

* [codex] Add Benchling workspace spending limits ([d1d4a6f](https://github.com/boltz-bio/boltz-api-cli/commit/d1d4a6ffd2268802a77310f8549c6f45f05b1efd))


### Bug Fixes

* **temporal:** require running workers in readiness probes ([e606b6e](https://github.com/boltz-bio/boltz-api-cli/commit/e606b6e8815a441eacadbdb545ec4a6375274985))

## 0.18.4 (2026-05-15)

Full Changelog: [v0.18.3...v0.18.4](https://github.com/boltz-bio/boltz-api-cli/compare/v0.18.3...v0.18.4)

## 0.18.3 (2026-05-15)

Full Changelog: [v0.18.2...v0.18.3](https://github.com/boltz-bio/boltz-api-cli/compare/v0.18.2...v0.18.3)

### Bug Fixes

* **install:** support Windows PowerShell arch fallback ([c7146f3](https://github.com/boltz-bio/boltz-api-cli/commit/c7146f32d25d3d10263b1dd28d54bee8d1ca4ccc))

## 0.18.2 (2026-05-15)

Full Changelog: [v0.18.1...v0.18.2](https://github.com/boltz-bio/boltz-api-cli/compare/v0.18.1...v0.18.2)

### Chores

* **internal:** codegen related update ([aee68a4](https://github.com/boltz-bio/boltz-api-cli/commit/aee68a45f89257a2d553cff23269ca550a00cb19))

## 0.18.1 (2026-05-13)

Full Changelog: [v0.18.0...v0.18.1](https://github.com/boltz-bio/boltz-api-cli/compare/v0.18.0...v0.18.1)

### Bug Fixes

* **billing:** mark compute billing stale from webhooks ([4bf1700](https://github.com/boltz-bio/boltz-api-cli/commit/4bf1700a4dcb03b6abb1a0959add4911e17f76c1))

## 0.18.0 (2026-05-13)

Full Changelog: [v0.17.0...v0.18.0](https://github.com/boltz-bio/boltz-api-cli/compare/v0.17.0...v0.18.0)

### Features

* **compute:** filter workspaces by name prefix ([5f834bf](https://github.com/boltz-bio/boltz-api-cli/commit/5f834bf62f069ffec1bca9c1ff6323c6fa4b41ab))

## 0.17.0 (2026-05-13)

Full Changelog: [v0.16.2...v0.17.0](https://github.com/boltz-bio/boltz-api-cli/compare/v0.16.2...v0.17.0)

### Features

* **admin:** add Compute API Monitor ([a52eaef](https://github.com/boltz-bio/boltz-api-cli/commit/a52eaefe6486e03b2ec07e14d684fd18f26f5ff5))

## 0.16.2 (2026-05-13)

Full Changelog: [v0.16.1...v0.16.2](https://github.com/boltz-bio/boltz-api-cli/compare/v0.16.1...v0.16.2)

## 0.16.1 (2026-05-12)

Full Changelog: [v0.16.0...v0.16.1](https://github.com/boltz-bio/boltz-api-cli/compare/v0.16.0...v0.16.1)

### Chores

* **internal:** codegen related update ([523e923](https://github.com/boltz-bio/boltz-api-cli/commit/523e923c62b297117254e119b308078d18c336cf))

## 0.16.0 (2026-05-11)

Full Changelog: [v0.15.7...v0.16.0](https://github.com/boltz-bio/boltz-api-cli/compare/v0.15.7...v0.16.0)

### Features

* [codex] Add ADME scoring for small molecule pipelines ([deaf796](https://github.com/boltz-bio/boltz-api-cli/commit/deaf796e5eed4f0ec39930999d7c7aeb024cc8d2))
* **compute-api:** support non-binding target residues ([75f1166](https://github.com/boltz-bio/boltz-api-cli/commit/75f1166c750db8519a908e7781c13dd03e7817ed))
* **download-results:** add pipeline download modes ([53639f2](https://github.com/boltz-bio/boltz-api-cli/commit/53639f22c00e53e89e720a73abc017b64343389e))


### Bug Fixes

* **download-results:** make default run names deterministic ([161ba33](https://github.com/boltz-bio/boltz-api-cli/commit/161ba339e5433b0b306708624944c937cca3791c))
* **download-results:** remove cif download mode ([3513eda](https://github.com/boltz-bio/boltz-api-cli/commit/3513eda88a4c55a4f8e56fa329e14fbc130e92cd))

## 0.15.7 (2026-05-11)

Full Changelog: [v0.15.6...v0.15.7](https://github.com/boltz-bio/boltz-api-cli/compare/v0.15.6...v0.15.7)

### Bug Fixes

* **download-results:** append pipeline manifest entries ([4f1284d](https://github.com/boltz-bio/boltz-api-cli/commit/4f1284d3b4881fde4529449053d596a0eaa1b117))
* **download-results:** surface manifest close errors ([ac5172e](https://github.com/boltz-bio/boltz-api-cli/commit/ac5172ed4757b79eabd3a6a2f7d74b4810cd4330))

## 0.15.6 (2026-05-09)

Full Changelog: [v0.15.5...v0.15.6](https://github.com/boltz-bio/boltz-api-cli/compare/v0.15.5...v0.15.6)

### Chores

* **internal:** codegen related update ([2145905](https://github.com/boltz-bio/boltz-api-cli/commit/2145905ad6425670704a4f115749ad4fe99a378a))

## 0.15.5 (2026-05-08)

Full Changelog: [v0.15.4...v0.15.5](https://github.com/boltz-bio/boltz-api-cli/compare/v0.15.4...v0.15.5)

### Chores

* redact api-key headers in debug logs ([1372f16](https://github.com/boltz-bio/boltz-api-cli/commit/1372f160ed4e6372bfa249695b386cf9383db09f))

## 0.15.4 (2026-05-04)

Full Changelog: [v0.15.3...v0.15.4](https://github.com/boltz-bio/boltz-api-cli/compare/v0.15.3...v0.15.4)

### Bug Fixes

* **auth:** use BOLTZ_API_KEY in auth guidance ([e772221](https://github.com/boltz-bio/boltz-api-cli/commit/e77222198ca2e4c8335a7b7161b39c3a88ccce7c))

## 0.15.3 (2026-05-02)

Full Changelog: [v0.15.2...v0.15.3](https://github.com/boltz-bio/boltz-api-cli/compare/v0.15.2...v0.15.3)

### Bug Fixes

* stainless changes ([85b5173](https://github.com/boltz-bio/boltz-api-cli/commit/85b5173bb90b0ad349ad3916fdf34de464a207d1))


### Chores

* update SDK settings ([c1f1e14](https://github.com/boltz-bio/boltz-api-cli/commit/c1f1e14ee63d73a7a87011061a9da3697bd0c92c))
* update SDK settings ([c887119](https://github.com/boltz-bio/boltz-api-cli/commit/c88711923fe0eb3959861c6255cfe5f84b3c677e))
* update SDK settings ([84a5c4e](https://github.com/boltz-bio/boltz-api-cli/commit/84a5c4e373fa19963567d25a9d57939a37ece32e))

## 0.15.2 (2026-05-02)

Full Changelog: [v0.15.1...v0.15.2](https://github.com/boltz-bio/boltz-compute-api-cli/compare/v0.15.1...v0.15.2)

### Chores

* **internal:** codegen related update ([d7ee403](https://github.com/boltz-bio/boltz-compute-api-cli/commit/d7ee4039cee1fe070b6421e2e8c72b04b7070728))

## 0.15.1 (2026-05-01)

Full Changelog: [v0.15.0...v0.15.1](https://github.com/boltz-bio/boltz-compute-api-cli/compare/v0.15.0...v0.15.1)

### Bug Fixes

* **compute-api:** polish reference docs for CLI and auth ([9c5063c](https://github.com/boltz-bio/boltz-compute-api-cli/commit/9c5063cb1f9f70da04957ad57be7ed3a7739150c))

## 0.15.0 (2026-05-01)

Full Changelog: [v0.14.1...v0.15.0](https://github.com/boltz-bio/boltz-compute-api-cli/compare/v0.14.1...v0.15.0)

### Features

* **compute:** add console usage page ([39373b7](https://github.com/boltz-bio/boltz-compute-api-cli/commit/39373b7da1dcb87309bee09b54193204729b0fbc))

## 0.14.1 (2026-05-01)

Full Changelog: [v0.14.0...v0.14.1](https://github.com/boltz-bio/boltz-compute-api-cli/compare/v0.14.0...v0.14.1)

### Chores

* **internal:** codegen related update ([93fb8ed](https://github.com/boltz-bio/boltz-compute-api-cli/commit/93fb8ed07cfca815aee5f40f1f92617b626a61d0))

## 0.14.0 (2026-04-30)

Full Changelog: [v0.13.0...v0.14.0](https://github.com/boltz-bio/boltz-compute-api-cli/compare/v0.13.0...v0.14.0)

### Features

* **compute:** gate console backend by compute roles ([c9abf9e](https://github.com/boltz-bio/boltz-compute-api-cli/commit/c9abf9e2d9b0268bfb3a0a0873d2f7ba636f6459))

## 0.13.0 (2026-04-30)

Full Changelog: [v0.12.0...v0.13.0](https://github.com/boltz-bio/boltz-compute-api-cli/compare/v0.12.0...v0.13.0)

### Features

* support passing path and query params over stdin ([50e8956](https://github.com/boltz-bio/boltz-compute-api-cli/commit/50e895648cdbad77f2b62ea9b7af837bdcdf2d34))


### Bug Fixes

* **sdk:** serialize usage arrays as repeated params ([80d9e48](https://github.com/boltz-bio/boltz-compute-api-cli/commit/80d9e48824ad14af63d908a64421a195af8e3543))

## 0.12.0 (2026-04-30)

Full Changelog: [v0.11.1...v0.12.0](https://github.com/boltz-bio/boltz-compute-api-cli/compare/v0.11.1...v0.12.0)

### Features

* add compute api retention settings ([3aebac6](https://github.com/boltz-bio/boltz-compute-api-cli/commit/3aebac6817fdcb4ada87407b7328ebdae83f3379))
* support passing path and query params over stdin ([854d8fb](https://github.com/boltz-bio/boltz-compute-api-cli/commit/854d8fb817e323ab1db2555bd87633b4cb506087))


### Bug Fixes

* **cli:** correctly load zsh autocompletion ([0970eb1](https://github.com/boltz-bio/boltz-compute-api-cli/commit/0970eb129385fdadb815b330ea78f2b92afc9bf4))
* flags for nullable body scalar fields are strictly typed ([067ef4d](https://github.com/boltz-bio/boltz-compute-api-cli/commit/067ef4da59dcafbda2944d6acb69298e8c408cde))

## 0.11.1 (2026-04-29)

Full Changelog: [v0.11.0...v0.11.1](https://github.com/boltz-bio/boltz-compute-api-cli/compare/v0.11.0...v0.11.1)

### Bug Fixes

* **auth:** extend OAuth callback wait ([231de6e](https://github.com/boltz-bio/boltz-compute-api-cli/commit/231de6eee570b9c51843d914822b68e989c063b4))

## 0.11.0 (2026-04-29)

Full Changelog: [v0.10.2...v0.11.0](https://github.com/boltz-bio/boltz-compute-api-cli/compare/v0.10.2...v0.11.0)

### Features

* **compute:** support curated binder specs ([f9b999e](https://github.com/boltz-bio/boltz-compute-api-cli/commit/f9b999e0acb889e9d0c04617cec6a21b9e2425cb))

## 0.10.2 (2026-04-29)

Full Changelog: [v0.10.1...v0.10.2](https://github.com/boltz-bio/boltz-compute-api-cli/compare/v0.10.1...v0.10.2)

### Bug Fixes

* **install:** make latest metadata updates monotonic ([063b5e3](https://github.com/boltz-bio/boltz-compute-api-cli/commit/063b5e35ad6e038e949feb6af13638bac08064bf))
* **install:** remove GitHub release fallback ([6a07330](https://github.com/boltz-bio/boltz-compute-api-cli/commit/6a073306e25ac7a14f4a13334728c60bc85a39aa))

## 0.10.1 (2026-04-28)

Full Changelog: [v0.10.0...v0.10.1](https://github.com/boltz-bio/boltz-compute-api-cli/compare/v0.10.0...v0.10.1)

### Bug Fixes

* **auth:** remove no-browser login option ([b7fdbb5](https://github.com/boltz-bio/boltz-compute-api-cli/commit/b7fdbb5425dc0203738aefecbf867029156f5098))

## 0.10.0 (2026-04-28)

Full Changelog: [v0.9.1...v0.10.0](https://github.com/boltz-bio/boltz-compute-api-cli/compare/v0.9.1...v0.10.0)

### Features

* **compute-api:** add CLI version metadata endpoint ([4f5cc87](https://github.com/boltz-bio/boltz-compute-api-cli/commit/4f5cc8743a328c0c3bd700a7b5437289bc5aece0))


### Bug Fixes

* revert generated SDK dependency bump ([957c84c](https://github.com/boltz-bio/boltz-compute-api-cli/commit/957c84cf4b0736cab4eb87c5c9a0a8b4fc5d6f2a))


### Chores

* **internal:** codegen related update ([21d8fb7](https://github.com/boltz-bio/boltz-compute-api-cli/commit/21d8fb7a2f1a2695ba48906b061a3a0666585473))

## 0.9.1 (2026-04-27)

Full Changelog: [v0.9.0...v0.9.1](https://github.com/boltz-bio/boltz-compute-api-cli/compare/v0.9.0...v0.9.1)

### Bug Fixes

* **cli:** harden install and config diagnostics ([f4f7663](https://github.com/boltz-bio/boltz-compute-api-cli/commit/f4f7663f6081733cd93b86f2d026a812f163ae2d))

## 0.9.0 (2026-04-27)

Full Changelog: [v0.8.1...v0.9.0](https://github.com/boltz-bio/boltz-compute-api-cli/compare/v0.8.1...v0.9.0)

### Features

* **compute-api:** add auth context endpoint ([0b89bd9](https://github.com/boltz-bio/boltz-compute-api-cli/commit/0b89bd98f8d1b5328f7f3958ad760e432f75efcb))
* **compute:** create workspace keys internally ([92ab298](https://github.com/boltz-bio/boltz-compute-api-cli/commit/92ab29867712c4538fa23c1729e4e3dba9538e5c))


### Bug Fixes

* **auth:** merge generated auth commands ([0c24e3f](https://github.com/boltz-bio/boltz-compute-api-cli/commit/0c24e3f616ea06d4efcb31a9375f88baa9ce401f))
* **stainless:** expose auth context as auth me ([4ada848](https://github.com/boltz-bio/boltz-compute-api-cli/commit/4ada848f3d52928a02cfe9728ec99062582a0183))

## 0.8.1 (2026-04-26)

Full Changelog: [v0.8.0...v0.8.1](https://github.com/boltz-bio/boltz-compute-api-cli/compare/v0.8.0...v0.8.1)

### Chores

* **internal:** version bump ([384fdd4](https://github.com/boltz-bio/boltz-compute-api-cli/commit/384fdd47c7cf8852354dc73b20e137e00b9b89f2))
* **internal:** version bump ([a513c20](https://github.com/boltz-bio/boltz-compute-api-cli/commit/a513c20e91df2f2d387e8cc8e6b491685ed238bb))


### Documentation

* **cli:** add cross-platform installer commands ([fc6a28a](https://github.com/boltz-bio/boltz-compute-api-cli/commit/fc6a28a1d98499f10506abfdecc7541454847156))

## 0.8.0 (2026-04-26)

Full Changelog: [v0.7.5...v0.8.0](https://github.com/boltz-bio/boltz-compute-api-cli/compare/v0.7.5...v0.8.0)

### Features

* **compute-api:** prefix structure-binding ids with sab_pred ([1c60be9](https://github.com/boltz-bio/boltz-compute-api-cli/commit/1c60be991d4174ecfdc580953c8634c618fd8598))


### Chores

* **internal:** codegen related update ([3e7dcc9](https://github.com/boltz-bio/boltz-compute-api-cli/commit/3e7dcc9863323d02fc0df06acb63b07c485ea6cb))
* **internal:** more robust bootstrap script ([a356ec3](https://github.com/boltz-bio/boltz-compute-api-cli/commit/a356ec3797a695168ffb9e3b174fe8da8efe860f))

## 0.7.5 (2026-04-24)

Full Changelog: [v0.7.4...v0.7.5](https://github.com/boltz-bio/boltz-compute-api-cli/compare/v0.7.4...v0.7.5)

## 0.7.4 (2026-04-23)

Full Changelog: [v0.7.3...v0.7.4](https://github.com/boltz-bio/boltz-compute-api-cli/compare/v0.7.3...v0.7.4)

## 0.7.3 (2026-04-23)

Full Changelog: [v0.7.2...v0.7.3](https://github.com/boltz-bio/boltz-compute-api-cli/compare/v0.7.2...v0.7.3)

### Bug Fixes

* **cli:** support sab prediction result downloads ([bc0fedf](https://github.com/boltz-bio/boltz-compute-api-cli/commit/bc0fedfeaf8df6933dc042115678fb048067920a))

## 0.7.2 (2026-04-23)

Full Changelog: [v0.7.1...v0.7.2](https://github.com/boltz-bio/boltz-compute-api-cli/compare/v0.7.1...v0.7.2)

## 0.7.1 (2026-04-22)

Full Changelog: [v0.7.0...v0.7.1](https://github.com/boltz-bio/boltz-compute-api-cli/compare/v0.7.0...v0.7.1)

### Bug Fixes

* **download-results:** reuse existing run dirs ([e2b52ff](https://github.com/boltz-bio/boltz-compute-api-cli/commit/e2b52ffe4380a9b184f04f050e598c92c8757d27))

## 0.7.0 (2026-04-22)

Full Changelog: [v0.6.0...v0.7.0](https://github.com/boltz-bio/boltz-compute-api-cli/compare/v0.6.0...v0.7.0)

### Features

* **cli:** add download status and jsonl progress ([97fcbf8](https://github.com/boltz-bio/boltz-compute-api-cli/commit/97fcbf8abb9e22bfce1ad3611debb49d1300facd))
* **cli:** support merged input payloads ([bcc2d31](https://github.com/boltz-bio/boltz-compute-api-cli/commit/bcc2d314ad998e100257920d44860ab300603dcb))

## 0.6.0 (2026-04-22)

Full Changelog: [v0.5.4...v0.6.0](https://github.com/boltz-bio/boltz-compute-api-cli/compare/v0.5.4...v0.6.0)

### Features

* **auth:** align OAuth login with Boltz provider ([504fd08](https://github.com/boltz-bio/boltz-compute-api-cli/commit/504fd0899a5da4d011978e20f89213c19f68464f))

## 0.5.4 (2026-04-22)

Full Changelog: [v0.5.3...v0.5.4](https://github.com/boltz-bio/boltz-compute-api-cli/compare/v0.5.3...v0.5.4)

### Bug Fixes

* **cli:** drop legacy run and screen ids ([3d65b17](https://github.com/boltz-bio/boltz-compute-api-cli/commit/3d65b1755ef00f2663672def596eec1420886d29))
* **cli:** require --id for download-results ([6f266c7](https://github.com/boltz-bio/boltz-compute-api-cli/commit/6f266c7a4d8202f39822f7113f1bbda78b063f82))

## 0.5.3 (2026-04-22)

Full Changelog: [v0.5.2...v0.5.3](https://github.com/boltz-bio/boltz-compute-api-cli/compare/v0.5.2...v0.5.3)

### Bug Fixes

* **compute-api:** replace unsafe application schema ([6790940](https://github.com/boltz-bio/boltz-compute-api-cli/commit/679094016e8a844e8c298580dd02a617db337ec0))

## 0.5.2 (2026-04-22)

Full Changelog: [v0.5.1...v0.5.2](https://github.com/boltz-bio/boltz-compute-api-cli/compare/v0.5.1...v0.5.2)

### Bug Fixes

* **compute-api:** default omitted polymer modifications ([c7ee7b0](https://github.com/boltz-bio/boltz-compute-api-cli/commit/c7ee7b03c782c7ef8c5c24831889d363ce3ef091))


### Refactors

* **cli:** isolate custom command code ([7235b5f](https://github.com/boltz-bio/boltz-compute-api-cli/commit/7235b5fb67146b1f99f07112af038b167d9b2984))

## 0.5.1 (2026-04-21)

Full Changelog: [v0.5.0...v0.5.1](https://github.com/boltz-bio/boltz-compute-api-cli/compare/v0.5.0...v0.5.1)

## 0.5.0 (2026-04-21)

Full Changelog: [v0.4.0...v0.5.0](https://github.com/boltz-bio/boltz-compute-api-cli/compare/v0.4.0...v0.5.0)

### ⚠ BREAKING CHANGES

* **compute-api:** rename pipeline route params to id

### Features

* **cli:** add download-results command ([7f541b8](https://github.com/boltz-bio/boltz-compute-api-cli/commit/7f541b89f13e1cef44e706d0c7bb54a46108deca))
* **compute:** accept user OAuth bearer tokens ([7a7dd8e](https://github.com/boltz-bio/boltz-compute-api-cli/commit/7a7dd8e94ebb3620356d167045d8d871c4d7a342))


### Refactors

* **compute-api:** rename pipeline route params to id ([0973e7d](https://github.com/boltz-bio/boltz-compute-api-cli/commit/0973e7dfce74fb872edc3ed69348997f09738ab7))

## 0.4.0 (2026-04-21)

Full Changelog: [v0.3.0...v0.4.0](https://github.com/boltz-bio/boltz-compute-api-cli/compare/v0.3.0...v0.4.0)

### Features

* **auth:** add backend OAuth provider ([da95a5f](https://github.com/boltz-bio/boltz-compute-api-cli/commit/da95a5f359c187e70a00cc98ef0c1094a84b4134))

## 0.3.0 (2026-04-21)

Full Changelog: [v0.2.0...v0.3.0](https://github.com/boltz-bio/boltz-compute-api-cli/compare/v0.2.0...v0.3.0)

### Features

* **cli:** add OAuth auth commands and session management ([4a545a7](https://github.com/boltz-bio/boltz-compute-api-cli/commit/4a545a741ace600c9f9dc90b38f9fd415a375bad))

## 0.2.0 (2026-04-21)

Full Changelog: [v0.1.0...v0.2.0](https://github.com/boltz-bio/boltz-compute-api-cli/compare/v0.1.0...v0.2.0)

### Features

* **billing:** attribute usage by product category ([4aa3932](https://github.com/boltz-bio/boltz-compute-api-cli/commit/4aa3932adefdc89439a8f92ce34956ef45aed2d6))


### Bug Fixes

* **compute-api:** support SM target bonds and constraints ([e1ab433](https://github.com/boltz-bio/boltz-compute-api-cli/commit/e1ab433367fac3a266759346229ac406944c1816))
* use friendly Go SDK package name ([5855866](https://github.com/boltz-bio/boltz-compute-api-cli/commit/585586697ae5ad7e0e09f66d6d0060c154f05ba4))

## 0.1.0 (2026-04-20)

Full Changelog: [v0.0.1...v0.1.0](https://github.com/boltz-bio/boltz-compute-api-cli/compare/v0.0.1...v0.1.0)

### Features

* rename cli command ([cd941c4](https://github.com/boltz-bio/boltz-compute-api-cli/commit/cd941c4baf451dcaa5ca803abcc793f6410458ad))
