# Changelog

## [0.9.0](https://github.com/nuvlaedge/nuvlaedge-go/compare/v0.8.1...v0.9.0) (2024-06-05)


### Features

* Add job engine container support for unsupported native jobs ([#48](https://github.com/nuvlaedge/nuvlaedge-go/issues/48)) ([9621a3a](https://github.com/nuvlaedge/nuvlaedge-go/commit/9621a3a6bae50a81e4c23d918b95f28ca88efc8a))
* Add Kubernetes installer for NuvlaEdge (non nuvlaedge-go) ([#47](https://github.com/nuvlaedge/nuvlaedge-go/issues/47)) ([b8446d0](https://github.com/nuvlaedge/nuvlaedge-go/commit/b8446d03c1ca699930adc71e2fb18fcef08842be))

## [0.8.1](https://github.com/nuvlaedge/nuvlaedge-go/compare/v0.8.0...v0.8.1) (2024-05-30)


### Dependencies

* Update Nuvla api client to 0.7.2 ([41f9efa](https://github.com/nuvlaedge/nuvlaedge-go/commit/41f9efa64b7c4883978634e788f18c4bec53cfbf))


### Minor Changes

* Add Environmental variables parsing for compose and swarm deployments ([#46](https://github.com/nuvlaedge/nuvlaedge-go/issues/46)) ([a5340f6](https://github.com/nuvlaedge/nuvlaedge-go/commit/a5340f67c3aa71cb199ebed80c4b4d58e64195a1))


### Code Refactoring

* Clean up unused files, functions and variables. ([bfe34c0](https://github.com/nuvlaedge/nuvlaedge-go/commit/bfe34c0f9ae973d8c95a723f349b735177bd4878))

## [0.8.0](https://github.com/nuvlaedge/nuvlaedge-go/compare/v0.7.3...v0.8.0) (2024-05-29)


### Features

* add docker stack deploy operation and stack executor ([#34](https://github.com/nuvlaedge/nuvlaedge-go/issues/34)) ([b6d728c](https://github.com/nuvlaedge/nuvlaedge-go/commit/b6d728cca8a5adc4885bfc7661d4018fb1b6b593))
* Add update and stop operation to stack operation ([#35](https://github.com/nuvlaedge/nuvlaedge-go/issues/35)) ([a128819](https://github.com/nuvlaedge/nuvlaedge-go/commit/a128819f926ad1ec0a9a14eb46223c6005577a6d))


### Bug Fixes

* Commit go.sum ([2efe412](https://github.com/nuvlaedge/nuvlaedge-go/commit/2efe41214b353d651ba76ee1ba74dd369eaa49a2))
* Fix update deployment for compose applications ([269722a](https://github.com/nuvlaedge/nuvlaedge-go/commit/269722a50e0ee2601bf3ff50fe6559c9d5ee62d1))
* Invert release build process order ([3b4ea18](https://github.com/nuvlaedge/nuvlaedge-go/commit/3b4ea18c64c604eb79314415dfa7c59ad92dd642))
* Remove darwin build from release ([4e58701](https://github.com/nuvlaedge/nuvlaedge-go/commit/4e587017b8fb064e5a0962f0aaa0ffc3bf16d915))


### Dependencies

* Update fsevents to fix CI error ([ccf070f](https://github.com/nuvlaedge/nuvlaedge-go/commit/ccf070f2001f2c72d5f06181adfdb5740bb23483))
* Update fsevents to fix CI error ([392474b](https://github.com/nuvlaedge/nuvlaedge-go/commit/392474b6f4d6974b56167b2afe96edd68d7f6306))


### Minor Changes

* Add service parameter for stack deployments ([60f0bdf](https://github.com/nuvlaedge/nuvlaedge-go/commit/60f0bdf1febf6e2a29753830787abbc17c1516df))


### Documentation

* Add MacOS usage documentation ([#29](https://github.com/nuvlaedge/nuvlaedge-go/issues/29)) ([b482f54](https://github.com/nuvlaedge/nuvlaedge-go/commit/b482f54dbeca046200c647f71ad7fed8d9866e5a))
* update usage ([56728d3](https://github.com/nuvlaedge/nuvlaedge-go/commit/56728d39160227f7642cc24923d5cc66fc5ab928))


### Code Refactoring

* Re-write job processor to allow multiple ways of executing actions with ease ([#33](https://github.com/nuvlaedge/nuvlaedge-go/issues/33)) ([61de4f9](https://github.com/nuvlaedge/nuvlaedge-go/commit/61de4f92e511f28020c325980ebd11cf2172ef96))
* rename jobEngine package into jobs ([d4c4bce](https://github.com/nuvlaedge/nuvlaedge-go/commit/d4c4bce0a796914b2a55d1857c17fde1c242fae9))


### Continuous Integration

* Add minor changes section to ChangLog notes ([5d3604a](https://github.com/nuvlaedge/nuvlaedge-go/commit/5d3604a0db1428fd8327ffc9b4c71ee78a067372))

## [0.7.3](https://github.com/nuvlaedge/nuvlaedge-go/compare/v0.7.2...v0.7.3) (2024-04-30)


### Bug Fixes

* bug on compressing binaries for macos ([121b3f5](https://github.com/nuvlaedge/nuvlaedge-go/commit/121b3f51831092a5f91d0d8a29e208d871483616))


### Continuous Integration

* fix ci release ([d55b5c0](https://github.com/nuvlaedge/nuvlaedge-go/commit/d55b5c08be3abccda1734868ef053204eb5c441a))

## [0.7.2](https://github.com/nuvlaedge/nuvlaedge-go/compare/v0.7.1...v0.7.2) (2024-04-30)


### Bug Fixes

* add compression to binaries release ([4578503](https://github.com/nuvlaedge/nuvlaedge-go/commit/4578503d916ece454033b3f12d6b70a3f67780c0))
* enable the service after installing, not only starting it ([fc8627f](https://github.com/nuvlaedge/nuvlaedge-go/commit/fc8627f86ab464d4c05f21b417ece7acd7313676))
* new get-ne.sh to point to nuvlaedge-cli installer ([507cdf1](https://github.com/nuvlaedge/nuvlaedge-go/commit/507cdf1842eaa8c5a48f2a30f3ada68dd35557ff))

## [0.7.1](https://github.com/nuvlaedge/nuvlaedge-go/compare/v0.7.0...v0.7.1) (2024-04-30)


### Bug Fixes

* fix release process bug ([ad7ae61](https://github.com/nuvlaedge/nuvlaedge-go/commit/ad7ae61ce332e2dfc1674c1008f1b8adca8a79b5))

## [0.7.0](https://github.com/nuvlaedge/nuvlaedge-go/compare/v0.6.0...v0.7.0) (2024-04-30)


### Features

* add cli installer ([#22](https://github.com/nuvlaedge/nuvlaedge-go/issues/22)) ([a1475a5](https://github.com/nuvlaedge/nuvlaedge-go/commit/a1475a53153827a17190d440db0807e3d8adf264))


### Bug Fixes

* bug on path composition for session freeze ([c97e980](https://github.com/nuvlaedge/nuvlaedge-go/commit/c97e98088bbfa4a464221856d393d3252bb0c2d5))
* release process version parsing bug ([1abb39e](https://github.com/nuvlaedge/nuvlaedge-go/commit/1abb39e2bfe7b27efaa46ad34c65cf541ba79d50))

## [0.6.0](https://github.com/nuvlaedge/nuvlaedge-go/compare/v0.5.0...v0.6.0) (2024-04-29)


### Features

* add persistence to nuvla session ([fec6cbb](https://github.com/nuvlaedge/nuvlaedge-go/commit/fec6cbb7fd47403671ddd982563d78b43573979a))

## [0.5.0](https://github.com/nuvlaedge/nuvlaedge-go/compare/v0.4.0...v0.5.0) (2024-04-26)


### Features

* add session persistence ([#18](https://github.com/nuvlaedge/nuvlaedge-go/issues/18)) ([971094a](https://github.com/nuvlaedge/nuvlaedge-go/commit/971094acfb786ce2dc203925e79de2b11672760e))

## [0.4.0](https://github.com/nuvlaedge/nuvlaedge-go/compare/v0.3.1...v0.4.0) (2024-04-25)


### Features

* add version management ([1a33237](https://github.com/nuvlaedge/nuvlaedge-go/commit/1a33237bed6b9f1517aaaf9456857ce15e3898a1))

## [0.3.1](https://github.com/nuvlaedge/nuvlaedge-go/compare/v0.3.0...v0.3.1) (2024-04-25)


### Bug Fixes

* remove replace dev statement from go.mod ([eaee032](https://github.com/nuvlaedge/nuvlaedge-go/commit/eaee032a7cbfca0a44c1b13e1fd8f42949d86e4d))

## [0.3.0](https://github.com/nuvlaedge/nuvlaedge-go/compare/v0.2.0...v0.3.0) (2024-04-25)


### Features

* add installer files to release ([29aad62](https://github.com/nuvlaedge/nuvlaedge-go/commit/29aad62491ad8bf3d327e5e7ca992caf26feff69))
* add script installer ([5c83cc0](https://github.com/nuvlaedge/nuvlaedge-go/commit/5c83cc0dc6bca44125558a86013d9061206b145d))
* add stop and improve deployment handling ([#15](https://github.com/nuvlaedge/nuvlaedge-go/issues/15)) ([4717dd4](https://github.com/nuvlaedge/nuvlaedge-go/commit/4717dd4a0bd9f59f7097837aa8595ff1b4371b2d))


### Bug Fixes

* error on exporting Settings path on installer ([3ea5c2f](https://github.com/nuvlaedge/nuvlaedge-go/commit/3ea5c2f4167945dee4f260d12e36d48e8685b7bc))
* fix default path for NuvlaEdge configuration in sudo mode ([e79a800](https://github.com/nuvlaedge/nuvlaedge-go/commit/e79a80069e9dd21f7e3ac16c37b7165171c03475))
* improve detached run mode on installer script ([5043c51](https://github.com/nuvlaedge/nuvlaedge-go/commit/5043c51913d49b3b538262c59909ca4595cc92ca))


### Dependencies

* updated api-client-go to 0.4.1 ([d99d22c](https://github.com/nuvlaedge/nuvlaedge-go/commit/d99d22ca5406458d528e59091ec2ab5203d22b23))

## [0.2.0](https://github.com/nuvlaedge/nuvlaedge-go/compare/v0.1.1...v0.2.0) (2024-03-26)


### Features

* Add deployment set state capabilities ([#11](https://github.com/nuvlaedge/nuvlaedge-go/issues/11)) ([10e97ff](https://github.com/nuvlaedge/nuvlaedge-go/commit/10e97ff1d85083235c0709a4773908cef015bc98))


### Dependencies

* Update client to 0.4.0 ([a54c4b2](https://github.com/nuvlaedge/nuvlaedge-go/commit/a54c4b2058e4add88cfeed59da60f552122cba70))

## [0.1.1](https://github.com/nuvlaedge/nuvlaedge-go/compare/v0.1.0...v0.1.1) (2024-03-26)


### Bug Fixes

* Deployment client no working ([3014a2d](https://github.com/nuvlaedge/nuvlaedge-go/commit/3014a2d99b8702a7505a5bbb357e33a0b5bcef8e))


### Dependencies

* Updated api-client-go to 0.3.1 ([3014a2d](https://github.com/nuvlaedge/nuvlaedge-go/commit/3014a2d99b8702a7505a5bbb357e33a0b5bcef8e))

## [0.1.0](https://github.com/nuvlaedge/nuvlaedge-go/compare/v0.0.1...v0.1.0) (2024-03-26)


### Features

* Add deployment start action ([#8](https://github.com/nuvlaedge/nuvlaedge-go/issues/8)) ([dd67280](https://github.com/nuvlaedge/nuvlaedge-go/commit/dd672802a653b34ff06c4f66e70f8d3abec9b2c8))


### Bug Fixes

* Remove Replace statement from go.mod ([641339c](https://github.com/nuvlaedge/nuvlaedge-go/commit/641339c38e0d469f71d1b5368568ebcc153e1bd8))


### Continuous Integration

* Add dependency section to release notes ([84af2a9](https://github.com/nuvlaedge/nuvlaedge-go/commit/84af2a9e68eb279e12cad57bd6a16fab33b1a9ac))
* Fix changelog sections ([b1f2e29](https://github.com/nuvlaedge/nuvlaedge-go/commit/b1f2e2991b21b000a7182e71b94a18010b5b5e58))
* Remove hardcoded initial release tag ([1a1f82a](https://github.com/nuvlaedge/nuvlaedge-go/commit/1a1f82a1311aff10b7797bb5a105b217534491b1))

## [0.0.1](https://github.com/nuvlaedge/nuvlaedge-go/compare/v0.0.1...v0.0.1) (2024-03-25)


### Features

* add release action ([#4](https://github.com/nuvlaedge/nuvlaedge-go/issues/4)) ([68a67e3](https://github.com/nuvlaedge/nuvlaedge-go/commit/68a67e3752eaaeb7f2f625dd8f7b37f05855688e))
* add-deployment actions skeletons ([2922de5](https://github.com/nuvlaedge/nuvlaedge-go/commit/2922de5e328fc163f009b0aaa674c192b7b1986b))
* add-reboot-action ([#2](https://github.com/nuvlaedge/nuvlaedge-go/issues/2)) ([ad2bf02](https://github.com/nuvlaedge/nuvlaedge-go/commit/ad2bf022370c54ef1d898cd9ae8bd3e72b036213))
* Added client library support ([653d426](https://github.com/nuvlaedge/nuvlaedge-go/commit/653d426cf95a76132d6150fbce95b77e79cfc542))


### Bug Fixes

* Simplify logging using base logger everywhere and clean of sourcecode ([f872a9e](https://github.com/nuvlaedge/nuvlaedge-go/commit/f872a9e23bf42bf9be5cd6403b84e9b710b7eac8))

## [0.1.0](https://github.com/nuvlaedge/nuvlaedge-go/compare/v0.0.1...v0.1.0) (2024-03-25)


### Features

* feat:  ([050d929](https://github.com/nuvlaedge/nuvlaedge-go/commit/050d92984f6c63c4976e157d2daf68f0908fbd8a))
* add-deployment actions skeletons ([2922de5](https://github.com/nuvlaedge/nuvlaedge-go/commit/2922de5e328fc163f009b0aaa674c192b7b1986b))
* add-reboot-action ([9e276dd](https://github.com/nuvlaedge/nuvlaedge-go/commit/9e276dd42e34bd4055cdc8057d2aa89b3ccd9fb0))
* add-reboot-action ([#2](https://github.com/nuvlaedge/nuvlaedge-go/issues/2)) ([ad2bf02](https://github.com/nuvlaedge/nuvlaedge-go/commit/ad2bf022370c54ef1d898cd9ae8bd3e72b036213))
* Added client library support ([653d426](https://github.com/nuvlaedge/nuvlaedge-go/commit/653d426cf95a76132d6150fbce95b77e79cfc542))


### Bug Fixes

* Simplify logging using base logger everywhere and clean of sourcecode ([f872a9e](https://github.com/nuvlaedge/nuvlaedge-go/commit/f872a9e23bf42bf9be5cd6403b84e9b710b7eac8))

## [0.2.0](https://github.com/nuvlaedge/nuvlaedge-go/compare/v0.1.0...v0.2.0) (2024-03-25)


### Features

* feat:  ([050d929](https://github.com/nuvlaedge/nuvlaedge-go/commit/050d92984f6c63c4976e157d2daf68f0908fbd8a))
* add-deployment actions skeletons ([2922de5](https://github.com/nuvlaedge/nuvlaedge-go/commit/2922de5e328fc163f009b0aaa674c192b7b1986b))
* add-reboot-action ([#2](https://github.com/nuvlaedge/nuvlaedge-go/issues/2)) ([ad2bf02](https://github.com/nuvlaedge/nuvlaedge-go/commit/ad2bf022370c54ef1d898cd9ae8bd3e72b036213))
* Added client library support ([653d426](https://github.com/nuvlaedge/nuvlaedge-go/commit/653d426cf95a76132d6150fbce95b77e79cfc542))


### Bug Fixes

* Simplify logging using base logger everywhere and clean of sourcecode ([f872a9e](https://github.com/nuvlaedge/nuvlaedge-go/commit/f872a9e23bf42bf9be5cd6403b84e9b710b7eac8))
