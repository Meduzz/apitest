# apitest
If  you're already using .http or .rest files for testing your api (supported via plugin, in most modern editors), this cli turns them into fully fledged automated tests.

`apitest infile.http`

### "How?" you ask...

First time you run this tool on a .http file, it will record the response (a distilled version). The following times you run the tool, it will compare what it gets back, with what was previously recorded. The recorded responses are clear text and easily readable. Remove things from the responses you dont care to test for.

### What if I add a new test?

The tool will recognize that there were a new test ran, and append the response to the .facit file. And the next time it will compare with what was stored.

## Install

* Install golang if you havent already.
* `go install github.com/Meduzz/apitest`
* ...and you're ready go :)

## Roadmap

* The [VS code extention](https://marketplace.visualstudio.com/items?itemName=humao.rest-client) for .http & .rest files supports parameters. This tool should shadow these capabilities on both ends (at least to some extent).
* There are files in the source code for 2 more powerful commands, `mock` and `record`. These should idealy be implemented.
* Coloured responses would be cool, but this is clearly a good to have.