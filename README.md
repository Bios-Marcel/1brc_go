My attempt at the billion rows challenge.

## Baseline

As a baseline, I tried to find a fast version written in Golang.

I went for this one for now: https://r2p.dev/b/2024-03-18-1brc-go/

I commented out the profiling code and made the path configurable from outside.

However, for me it takes way longer than 1.96 seconds, as I have pretty bad
hardware in comparison.

## My approach

I am trying to keep it as simple as I can. The main difference is how I approach
file reading. Instead of having one routine that reads, I am making use of the
fact that SSDs don't have a spinning head anymore. I open multiple files and
seek different locations. This allows me to have blocking free parallel reads.

Additionally, ints are parsed manually and we attempt to do calculations as late
as we can, unless not feasible.

## Run benchmarks

Simply execute `benchmark.ps1 measurements.txt` or make a bash variant of it.

You can pass different sizes of measurement files, however, as per the rules,
the file is optimised for 1 billion and might even fail on small files!

