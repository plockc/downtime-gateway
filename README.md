## July 4, 2023

Considering heirarcy of data, like Namespace -> IPSet -> IPSetIP

Each has an Id

IPSets know they have IPs, and Namespace knows it has IPSets.

Functions go on the IP but function needs IPSet, Namespace, etc.

Embed the Parent (inverted from Class heirarchy)

IP {
  Id string
  IPSet {
    Id string
    Namespace {
      Id string
    }
  }
}

Handlers should be independant generic overlay on top of a structure, with an adapter layer
So, all the structs above should impmlement a Resource interface that the handlers can work with (CRUD)


URL looks like /namespace/foo/ipset/bar/ip/012034200 for specifying an IP address

Create namespace foo first

Then find a MemberFactory keyed by "ipset", which can take the parent resource and provide a factory for IPSets

That factory creates an IPSet with the Namespace and recurse this process to create IP.




