# buildnumbers

Internal tool that keeps track of and hands out build numbers.

For the versioning, we are roughly following [this scheme](https://tidbits.com/2020/07/08/how-to-decode-apple-version-and-build-numbers/). E.g., `0.3.0_0C123` = Version `0.3.0`, build number `0C123` (123rd build in the 0.3 series). Note that prerelease builds may already have the version number of the upcoming release. 

## Displaying and interpreting build numbers

This system is there because many builds may be needed until a release is made to the public under a public-facing version ("marketing vertsion").

Example:

* 0.2.0 = the current helloSystem release (a.k.a. the "marketing version")
* 0.3.0 = the upcoming release; all continuous/experimental builds are builds toward that release
* 0C139 = the 139st build for what will become released as the 0.3.x series (a.k.a. the "build number")

In the FreeBSD package, we need to use  `0.3.0_0C139` due to restrictions of FreeBSD tooling but `0.3.0 (0C139)` is the correct way to display this in user interfaces so that it becomes clear that the build number is not part of the marketing version.

## This tool

This tool hands out the build numbers to Cirrus CI roughly.

In order to do this, it needs to persist the last handed out build number per minor version somewhere if the `PERSIST_NEW_BUILDNUMBER` environment variable is set to `YES` (this should be done for _one_ of the builds in a build matrix, so that all builds of that matrix get the same build number).
For this, the description field of the GitHub Release (described by `RELEASE_ID_FOR_STORAGE`)
is used as the storage to keep track of which build numbers have already been handed out.
