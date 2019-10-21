# podsaver
an app to download episodes of your favourite podcast to your local archive

##### Table of Contents
* [Why](#why)
* [Usage](#usage)
* [How it works](#how-it-works)
* [Building the app](#building-the-app)
* [Privacy considerations](#privacy-considerations)
* [Credits](#credits)

## Why
Sometimes things go offline for good, so some of us want a local archive of a favourite podcast to re-listen to during long winter nights.

## Usage
```
podsaver [-r] [-p path] url
```
The app will process the podcast feed (RSS or Atom) at a given `url`, scan the existing downloaded episodes in `path` (current directory by default) and try to download new and/or missing episodes.

## How it works
`podsaver` scans the local directory (current working directory if `-p path` is not given) and tries to guess podcast episode numbers from filenames (i.e. `episode14.mp3` or `14-new-guitar.avi` will be detected as episode 14 of the podcast). It then fetches the podcast feed from the given `url` and tries to match episodes in the feed to the latest one in the local directory. If matched, `podsaver` will download any episodes listed in the feed that are missing locally and save them as `number.ext` file (`ext` being extension such as `mp3` or `avi`) into the local directory. If given `-r` option, `podsaver` will also rename the existing files in a consistent manner.

For example, you have the following files:
```
episode3.mp3
episode 5.mp3
6.mp3
10.mp3
11-new-songs.mp3
awesomepodcast.13.mp3
```
and the feed has episodes 6 to 15 listed in it. `podsaver` will download what it can, so the folder will have:
```
episode3.mp3
episode 5.mp3
6.mp3
07.mp3
08.mp3
09.mp3
10.mp3
11-new-songs.mp3
12.mp3
awesomepodcast.13.mp3
14.mp3
15.mp3
```
`podsaver` has no way to get episodes 1, 2 and 4, but what was in the feed has been downloaded. Notice how the app appended trailing zeros to episodes 7, 8 and 9, because the latest episodes are in double-digit numbers.


Had you used `-r` option, existing files would be renamed, too:
```
03.mp3
05.mp3
06.mp3
07.mp3
08.mp3
09.mp3
10.mp3
11.mp3
12.mp3
13.mp3
14.mp3
15.mp3
```

## Building the app
```
go run build.go
```

## Credits
This software includes the following software or parts thereof:
* [Afero](https://github.com/spf13/afero) Copyright © 2014 Steve Francia
* [gofeed](https://github.com/mmcdole/gofeed) Copyright © 2016 mmcdole
* [gorilla/feeds](https://github.com/gorilla/feeds) Copyright © 2013-2018 The Gorilla Feeds Authors
* [The Go Programming Language](https://golang.org) Copyright 2009 The Go Authors
