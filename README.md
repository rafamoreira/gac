# GAC

## Git auto commit.

Simple cli app to speed up commits for single user repositories.

Many times I'm working in a repository that nobody is going to ever read a
commit message, this is helping with that.

Right now the implementation is very simple, it just adds all files and commits
with the sequential number of the commit.

This repository uses the tool to handle the commits itself.

## Recommendation

Don't use this tool, specially if you are working in a team, this is just a
tool to speed up the commits for single user repositories.
And even for single user repositories, it's better to write a good commit
message, and it was developed just as an exercise to learn more about Go, and
AUR packages.
But if you want to use it anyway, be careful with your secrets and sensitive
information, this tool will add all files to the commit.

## Usage

```bash
gac
```

[![Go](https://github.com/rafamoreira/gac/actions/workflows/go.yml/badge.svg?branch=master)](https://github.com/rafamoreira/gac/actions/workflows/go.yml)