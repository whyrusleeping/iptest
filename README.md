# iptest

A very simple tool for testing 'ipfs add' with set configurations.

## Usage
```
Usage of ./iptest:
  -chunker string
        chunker to use while adding
  -dir-depth int
        random-files directory depth parameter (default 3)
  -file-size int
        random-files filesize parameter (default 100000)
  -ipfsbin string
        ipfs binary to use
  -num-dirs int
        random-files dirs parameter (default 5)
  -num-files int
        random-files number of files per dir (default 10)
  -raw-leaves
        use raw leaves for add
  -repo-sync
        flatfs datastore sync (default true)
  -routing string
        specify routing type to use
  -to-add string
        optionally specify data to test adding
```

## License
MIT
