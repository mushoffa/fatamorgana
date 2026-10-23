# Changelog

All notable changes to this project will be documented in this file. See [standard-version](https://github.com/conventional-changelog/standard-version) for commit guidelines.

## [0.7.0](https://github.com/mushoffa/fatamorgana/compare/v0.6.0...v0.7.0) (2026-10-23)


### Breaking Changes

* **histogram:** Histogram.HORIZONTAL and histogram.VERTICAL are now
histogram.Horizontal and histogram.Vertical. ([415c463](https://github.com/mushoffa/fatamorgana/commit/415c463668431650ead372f32f8beed1a80a9060))



### Features

* **histogram:** **BREAKING** Add Binary profile and rename axis constants ([415c463](https://github.com/mushoffa/fatamorgana/commit/415c463668431650ead372f32f8beed1a80a9060))
* **threshold:** Add Sauvola adaptive thresholding ([5989743](https://github.com/mushoffa/fatamorgana/commit/59897439d270b8eeb8ce816600adc021b6e5ad37))


### Tests

* **histogram:** Add Binary tests and benchmark ([829a87f](https://github.com/mushoffa/fatamorgana/commit/829a87fcdec9cd14f99fbac2f78a6d70ee90d913))
* **threshold:** Add Otsu unit tests and benchmarks ([73f38e5](https://github.com/mushoffa/fatamorgana/commit/73f38e5d40706787047a95176f88be92dfd87f6c))
* **threshold:** Add Sauvola unit tests and benchmark ([bdafb12](https://github.com/mushoffa/fatamorgana/commit/bdafb124399a79073abbbc88ab659775f8940a2e))

## [0.6.0](https://github.com/mushoffa/fatamorgana/compare/v0.5.0...v0.6.0) (2026-10-19)



### Features

* **threshold:** Add Otsu global thresholding ([83b6e2a](https://github.com/mushoffa/fatamorgana/commit/83b6e2a3f2dd1af0d5055243b6ddbeac56e932e9))
* **filter:** Add Savitzky-Golay filter ([9ebfc14](https://github.com/mushoffa/fatamorgana/commit/9ebfc143fe1a4856b00e463e53284af07eea9411))
* Add Version() to report the module version ([7fd5e6f](https://github.com/mushoffa/fatamorgana/commit/7fd5e6fe5cf073eda5d8d60dc406247489ccf6e3))

## [0.5.0](https://github.com/mushoffa/fatamorgana/compare/v0.4.0...v0.5.0) (2025-03-07)


### Features

* **type:** Change JPEG string extension type to .jpeg ([f4718c9](https://github.com/mushoffa/fatamorgana/commit/f4718c97a5813bed973d119c12a2e6cda6029f44))

## [0.4.0](https://github.com/mushoffa/fatamorgana/compare/v0.3.0...v0.4.0) (2025-03-04)


### Features

* **crop:** Add image crop function ([13f3fc9](https://github.com/mushoffa/fatamorgana/commit/13f3fc996cf4c05411ec695afcaf38d742c45a37))
* **histogram:** Add generate histogram function for monochrome image ([8b5c66e](https://github.com/mushoffa/fatamorgana/commit/8b5c66ebd439cd987295a25548989ef201c7cf64))

## [0.3.0](https://github.com/mushoffa/fatamorgana/compare/v0.2.0...v0.3.0) (2025-03-03)


### Features

* **image:** Add new image with extension type function ([71b5799](https://github.com/mushoffa/fatamorgana/commit/71b5799d7c3da916563069c04cf71e6054ecde84))

## [0.2.0](https://github.com/mushoffa/fatamorgana/compare/v0.1.0...v0.2.0) (2025-03-03)


### Features

* **image:** Add new function to return image.Image type ([304711a](https://github.com/mushoffa/fatamorgana/commit/304711a30cb63c8fa9cb90168284d19cc09b051e))

## 0.1.0 (2025-03-03)


### Features

* Add Image struct and conversion function ([c8aa268](https://github.com/mushoffa/fatamorgana/commit/c8aa2685d9299d02f34ab9e7389d861c48e1fb3d))
* Add image type enum list ([c04c6cc](https://github.com/mushoffa/fatamorgana/commit/c04c6ccd1b64ef4d21ef1be75e34480a8cafcb3a))
* **decoder:** Add image decoder reader function ([9c21b3d](https://github.com/mushoffa/fatamorgana/commit/9c21b3d1acdd18749faa361eea8cb75358cc8993))
* **encoder:** Add image encoder function ([1c955f1](https://github.com/mushoffa/fatamorgana/commit/1c955f1fbd7481f46c109228bc0c70524f87ccf1))
* **file:** Add image read/write function ([5d0549b](https://github.com/mushoffa/fatamorgana/commit/5d0549b673f6c3ded9bc0e5ffc0ad45e96ed3bb6))
* **grayscale:** Add common grayscale conversion algorithm methods ([b8ed820](https://github.com/mushoffa/fatamorgana/commit/b8ed8205b7732b5a5d1d59c0ebad05723242c432))
* **image:** Add MimeType and Format function ([0370425](https://github.com/mushoffa/fatamorgana/commit/037042523ecdef36f44063deef792f3a38b01c6b))


### Documentation

* Add LICENSE file ([5dcb10c](https://github.com/mushoffa/fatamorgana/commit/5dcb10c3e3da1ddbd09acf6c88e0f27310cfa459))


### Build System

* Add go.mod file ([dffd5a9](https://github.com/mushoffa/fatamorgana/commit/dffd5a9f7b0159b86e1a80ca340911608684b5f1))
