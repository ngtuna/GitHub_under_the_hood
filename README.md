# GitHub - Under the hood

### Version Control System design
- storing content
- tracking changes to the content (history including merge metadata)
- distributing the content and history with collaborators

### Content storage
The most common design choices for storing content in the VCS world are with a delta-based changeset, or with directed acyclic graph (DAG) content representation.

Delta-based changesets encapsulate the differences between two versions of the flattened content, plus some metadata. Representing content as a directed acyclic graph involves objects forming a hierarchy which mirrors the content's filesystem tree as a snopshot of the commit (reusing the unchanged objects inside the tree where possible). Git stores content as a directed acyclic graph using different types of objects. The "Object Database" section later in this chapter describes the different types of objects that can form DAGs inside the Git repository.

### Object Model
#### SHA
All the information needed to represent the history of a project is stored in files referenced by a 40-digit "object name" that looks something like this:
```
6ff87c4664981e4397625791c8ea3bbb5f2279a3
```
You will see these 40-character strings all over the place in Git. In each case the name is calculated by taking the SHA1 hash of the contents of the object. The SHA1 hash is a cryptographic hash function. What that means to us is that it is virtually impossible to find two different objects with the same name.

#### Objects
Every object consists of three things - a **type**, a **size** and **content**. The `size` is simply the size of the contents, the contents depend on what type of object it is, and there are four different types of objects: `blob`, `tree`, `commit`, and `tag`.

- A `blob` is used to store file data - it is generally a file.
- A `tree` is basically like a directory - it references a bunch of other `trees` and/or `blobs` (i.e. files and sub-directories)
- A `commit` points to a single tree, marking it as what the project looked like at a certain point in time. It contains meta-information about that point in time, such as a timestamp, the author of the changes since the last commit, a pointer to the previous commit(s), etc.
- A `tag` is a way to mark a specific commit as special in some way. It is normally used to tag certain commits as specific releases or something along those lines.

Almost all of Git is built around manipulating this simple structure of four different object types. It is sort of its own little filesystem that sits on top of your machine's filesystem.

##### Different from SVN
It is important to note that this is very different from most SCM systems that you may be familiar with. Subversion, CVS, Perforce, Mercurial and the like all use Delta Storage systems - they store the differences between one commit and the next. Git does not do this - it stores a snapshot of what all the files in your project look like in this tree structure each time you commit. This is a very important concept to understand when using Git.

##### Blob Object
A `blob` generally stores the contents of a file.

![Alt text](/images/object-blob.png?raw=true)

A `blob` object is nothing but a chunk of binary data. It doesn't refer to anything else or have attributes of any kind, not even a file name.

Since the `blob` is entirely defined by its data, if two files in a directory tree (or in multiple different versions of the repository) have the same contents, they will share the same `blob` object. The object is totally independent of its location in the directory tree, and renaming a file does not change the object that file is associated with.

##### Tree Object
A `tree` is a simple object that has a bunch of pointers to `blobs` and other `trees` - it generally represents the contents of a directory or subdirectory.

![Alt text](/images/object-tree.png?raw=true)

As you can see, a `tree` object contains a list of entries, each with a mode, object type, SHA1 name, and name, sorted by name. It represents the contents of a single directory tree.

An object referenced by a `tree` may be blob, representing the contents of a file, or another `tree`, representing the contents of a subdirectory. Since trees and blobs, like all other objects, are named by the SHA1 hash of their contents, two trees have the same SHA1 name if and only if their contents (including, recursively, the contents of all subdirectories) are identical. This allows git to quickly determine the differences between two related `tree` objects, since it can ignore any entries with identical object names.

##### Commit Object
The `commit` object links a physical state of a `tree` with a description of how we got there and why.

![Alt text](/images/object-commit.png?raw=true)

As you can see, a `commit` is defined by:
- a `tree`: The SHA1 name of a `tree` object (as defined below), representing the contents of a directory at a certain point in time.
- **parent(s)**: The SHA1 name of some number of `commits` which represent the immediately previous step(s) in the history of the project. The example above has one parent; merge `commits` may have more than one. A `commit` with no parents is called a "root" commit, and represents the initial revision of a project. Each project must have at least one root. A project can also have multiple roots, though that isn't common (or necessarily a good idea).
- an **author**: The name of the person responsible for this change, together with its date.
- a **committer**: The name of the person who actually created the `commit`, with the date it was done. This may be different from the author; for example, if the author wrote a patch and emailed it to another person who used the patch to create the `commit`.
- a **comment**: describing this commit.

Note that a `commit` does not itself contain any information about what actually changed; all changes are calculated by comparing the contents of the `tree` referred to by this `commit` with the `trees` associated with its parents. In particular, git does not attempt to record file renames explicitly, though it can identify cases where the existence of the same file data at changing paths suggests a rename.

##### The Object Model
So, now that we've looked at the 3 main object types (`blob`, `tree` and `commit`), let's take a quick look at how they all fit together.

If we had a simple project with the following directory structure:
```
$>tree
.
|-- README
`-- lib
    |-- inc
    |   `-- tricks.rb
    `-- mylib.rb

2 directories, 3 files
```
And we committed this to a Git repository, it would be represented like this:

![Alt text](/images/objects-example.png?raw=true)

You can see that we have created a `tree` object for each directory (including the root) and a `blob` object for each file. Then we have a `commit` object to point to the root, so we can track what our project looked like when it was committed.

##### Tag Object
A `tag` object contains an object name (called simply 'object'), object type, tag name, the name of the person ("tagger") who created the tag, and a message, which may contain a signature.

![Alt text](/images/object-tag.png?raw=true)

#### Understand the .Git folder
It is the directory that stores all Git's history and meta information for your project - including all of the objects (commits, trees, blobs, tags), all of the pointers to where different branches are and more.

There is only one .Git directory per project. If you look at the contents of that directory, you can see all of important hidden files:
```
$ tree -L 1
.
|-- HEAD         # pointer to your current branch
|-- config       # your configuration preferences
|-- description  # description of your project 
|-- hooks/       # pre/post action hooks
|-- index        # index file (see next section)
|-- logs/        # a history of where your branches have been
|-- objects/     # your objects (commits, trees, blobs, tags)
|-- refs/        # pointers to your branches
```
#### How Git stores objects physically
All objects are stored as compressed contents by their sha values. They contain the object type, size and contents in a gzipped format.

There are two formats that Git keeps objects in - loose objects and packed objects.

##### Loose objects
Loose objects are the simpler format. It is simply the compressed data stored in a single file on disk. Every object written to a seperate file.

If the sha of your object is `ab04d884140f7b0cf8bbf86d6883869f16a46f65`, then the file will be stored in the following path:
```
GIT_DIR/objects/ab/04d884140f7b0cf8bbf86d6883869f16a46f65
```
It pulls the first two characters off and uses that as the subdirectory, so that there are never too many objects in one directory. The actual file name is the remaining 38 characters.

##### Packed objects
The other format for object storage is the packfile. Since Git stores each version of each file as a seperate object, it can get pretty inefficient. Imagine having a file several thousand lines long and changing a single line. Git will store the second file in it's entirety, which is a great big waste of space.

In order to save that space, Git utilizes the packfile. This is a format where Git will only save the part that has changed in the second file, with a pointer to the file it is similar to.

When objects are written to disk, it is often in the loose format, since that format is less expensive to access. However, eventually you'll want to save the space by packing up the objects - this is done with the `git gc` command. It will use a rather complicated heuristic to determine which files are likely most similar and base the deltas off that analysis. There can be multiple packfiles, they can be repacked if neccesary (`git repack`) or unpacked back into loose files (`git unpack-objects`) relatively easily.

Git will also write out an index file for each packfile that is much smaller and contains offsets into the packfile to more quickly find specific objects by sha.

The actual details of the packfile implementation are found in the Packfile chapter a little later on.

#### DEMO
A low-level process to make a new commit and push to remote branch is something like this:

**Create a new file and then generate a new `blob`**
```
$ git hash-object -w new-commit.txt 
```
**Create a new tree by adding the new generated blob**
- check the current top commit
```
$ cat .git/refs/heads/master 
```
- show the tree that the top commit is pointing to
```
$ git show -s --pretty=raw <sha of commit>
```
- show the content of the tree
```
$ git ls-tree <sha of tree>
```
- create a new tree by writing new content including the last blob to a temporary file `/tmp/tree.txt`
- then make a new tree
```
$ cat /tmp/tree.txt | git mktree
```
**Create a new commit pointed to the generated tree**
```
$ export GIT_AUTHOR_NAME=ngtuna
$ export GIT_AUTHOR_EMAIL=ng.tuna@gmail.com
$ export GIT_COMMITTER_NAME=ngtuna
$ export GIT_COMMITTER_EMAIL=ng.tuna@gmail.com
$ git commit-tree <sha of generated tree> -p <parent commit> -m "commit message" < /tmp/message
```
**Updating the Branch Ref**
```
$ git update-ref refs/heads/master <sha of new commit>
```
**Push to remote repo**
```
$ git push
```
