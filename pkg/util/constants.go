package util

const HelpDocDir string = "help_docs"
const HelpDocExtension = ".txt"
const JitVersion = "1.0.0"
const JitDirName = ".jit"

const MAIN = "main"
const HEAD = "head"
const STAGE = "stage"
const CONFIG = "config"
const LOGS = "logs"
const INFO = "info"
const BRANCHES = "branches"
const SNAPSHOTS = "snapshots"
const OBJECTS = "objects"

const DefaultFilePerm = 0644

const Init string = "init"

type File string

const DataFile File = "dataFile"
const Directory File = "directory"
