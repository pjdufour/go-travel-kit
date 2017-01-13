[![Build Status](https://travis-ci.org/pjdufour/go-extract.svg?branch=master)](https://travis-ci.org/pjdufour/go-extract)

# go-travel-kit

## Usage

Travel Kit is released as a single binary with static resources and html templates embedded using [go.rice](https://github.com/GeertJohan/go.rice), so no other runtime dependencies are needed.  You can specify configuration settings with a local YAML file or via environmental variables.  For the YAML file, simply pass it as the first parameter.

```
./travelkit travelkit.yml
```

For example, here is an example configuration file.

```
TRAVELKIT_HOME: /home/vagrant/.travelkit
SITE:
  NAME: "Patrick's Travel Kit"
  URL: http://localhost:8002
MEDIA:
  PAGE_SIZE: 100
  LOCATIONS:
    - "~/Desktop"
    - "~/workspaces"
    - "~/GIS"

```

You could also set some of the settings using inline enivornmental variables as below, which switches the port from the default `8000` to `8001`.

```
SITE_URL=http://localhost:8001 ./travelkit
```

## Building

Install all dependencies with

```shell
go get ./...
# or
go get github.com/pjdufour/go-gypsy/yaml  # fork of github.com/kylelemons/go-gypsy/yaml
```

**Build Static Files**

To build static files, cd into the project directory (e.g., `~/src/github.com/pjdufour/go-travel-kit`) and then run:

```
rice -v embed-go --import-path=./travelkit
```

Or in 1 line as:

```
cd ~/src/github.com/pjdufour/go-travel-kit/; rice -v embed-go --import-path=./travelkit
```

Rice requires static folders to be within the same folder

To build a package, cd back into home (e.g., `cd`), and then run

```shell
go build github.com/pjdufour/go-travel-kit/cmd/travelkit
```


## Testing

To run the tests

```shell
go test github.com/pjdufour/go-travel-kit/cmd/travelkit
```

## License

Copyright (c) 2017, Patrick Dufour
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

* Redistributions of source code must retain the above copyright notice, this
  list of conditions and the following disclaimer.

* Redistributions in binary form must reproduce the above copyright notice,
  this list of conditions and the following disclaimer in the documentation
  and/or other materials provided with the distribution.

* Neither the name of go-travel-kit nor the names of its
  contributors may be used to endorse or promote products derived from
  this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
