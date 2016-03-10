# GitHub - Under the hood

### Version Control System design
- storing content
- tracking changes to the content (history including merge metadata)
- distributing the content and history with collaborators

### Content storage
The most common design choices for storing content in the VCS world are with a delta-based changeset, or with directed acyclic graph (DAG) content representation.

Delta-based changesets encapsulate the differences between two versions of the flattened content, plus some metadata. Representing content as a directed acyclic graph involves objects forming a hierarchy which mirrors the content's filesystem tree as a snopshot of the commit (reusing the unchanged objects inside the tree where possible). Git stores content as a directed acyclic graph using different types of objects. The "Object Database" section later in this chapter describes the different types of objects that can form DAGs inside the Git repository.

