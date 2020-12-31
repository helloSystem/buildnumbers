# buildnumbers

Internal tool that keeps track of and hands out build numbers.

For the versioning, we are roughly following [this scheme](https://tidbits.com/2020/07/08/how-to-decode-apple-version-and-build-numbers/). E.g., `0.3.0_0C123` = Version `0.3.0`, build number `0C123` (123rd build in the 0.3 series). Note that prerelease builds may already have the version number of the upcoming release. 

This tool hands out the build numbers to Cirrus CI roughly.

In order to do this, it needs to persist the last handed out build number per minor version somewhere.
For this, the description field of the GitHub Release (described by `RELEASE_ID_FOR_STORAGE`)
is used as the storage to keep track of which build numbers have already been handed out.
