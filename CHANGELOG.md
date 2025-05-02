# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [[0.1.0](https://github.com/Endg4meZer0/lrcsnc/releases/tag/v0.1.0)] - 2025-05-03
### Added
- Some simple unit tests like cache and romanization.
- Makefile for easier build and more control over linking stuff.
- JSON piped output.
### Changed
- **Everything is rewrote from scratch**:
- MPRIS support now works on signals instead of polling.
- Configuration format is now TOML.
- Japanese romanization now uses [kakasi](https://github.com/loretoparisi/kakasi) and is able to romanize kanji. Kakasi is installed as a separate dependency for it to work.
- Romanization is now able to handle multiple languages on one line.
- Flags usage is now handled by [go-flags](https://github.com/jessevdk/go-flags) instead of standard library.
### Removed
- Playerctl dependency is now completely abandoned and cut from the app in favor of direct MPRIS handling using [own library](https://github.com/Endg4meZer0/go-mpris).
- Terminal output in one line is removed.

## The changelog below describes [playerctl-lyrics](https://github.com/Endg4meZer0/playerctl-lyrics), the now archived project which lrcsnc is based on.

## [[0.2.1](https://github.com/Endg4meZer0/playerctl-lyrics/releases/tag/v0.2.1)] - 2024-08-29
### Added
- A command-line option `-o` to redirect the output to a set file.
- ~~A command-line option to display lyrics in one line~~ **is removed now**
- A configuration option to offset the lyrics by set seconds by @Endg4meZer0 in [#9](https://github.com/Endg4meZer0/playerctl-lyrics/pull/9)
### Changed
- More refactoring: `cmus` and other players that report position in integer seconds are now fully supported.
- Cache system is reverted back to JSON instead of LRC files to allow more additional data to be stored ([#10](https://github.com/Endg4meZer0/playerctl-lyrics/pull/10))
### Fixed
- Instrumental lyrics overlapped actual lyrics in some cases ([#11](https://github.com/Endg4meZer0/playerctl-lyrics/pull/11))

## [[0.2.0](https://github.com/Endg4meZer0/playerctl-lyrics/releases/tag/v0.2.0)] - 2024-08-24
### Changed
- A big concept rewrite happened to allow players like `cmus` that report position in integer seconds work on par with others.
- A rename of `doCacheLyrics` configuration option to `enabled`

## [[0.1.1](https://github.com/Endg4meZer0/playerctl-lyrics/releases/tag/v0.1.1)] - 2024-08-21
### Added
- A configuration option to control the format of repeated lyrics multiplier.
### Fixed
- Fixed a panic if there is no space after a timestamp.
- Fixed a panic when romanization of Japanese kanji failed and fell down to Chinese characters 

## [[0.1.0](https://github.com/Endg4meZer0/playerctl-lyrics/releases/tag/v0.1.0)] - 2024-08-15
### Added
- Initial unstable release of playerctl-lyrics.
- Display lyrics for currently playing song.
- Support for multiple music players using `playerctl`.
- Automatic lyric fetching from `lrclib`.
- Configuration file for custom settings.
- Romanization for several asian languages.
- Caching system to significantly reduce traffic usage.