function buildFileTree(data) {
  const dirs = data.filter((x) => x.type === "dir");
  const files = data.filter((x) => x.type === "file");
  const cache = new Map();

  let rootDir = {
    id: "root",
    name: "root",
    parentId: undefined,
    type: "DIRECTORY",
    path: "",
    depth: 0,
    dirs: [],
    files: [],
  };

  dirs.forEach((item) => {
    let dir = {
      id: item.path,
      name: item.name,
      path: item.path,
      parentId:
        item.path.split("/").length === 2
          ? "0"
          : dirs.find(
              (x) => x.path === item.path.split("/").slice(0, -1).join("/")
            )?.path,
      type: "DIRECTORY",
      depth: 0,
      dirs: [],
      files: [],
    };
    cache.set(dir.id, dir);
  });

  files.forEach((item) => {
    let file = {
      id: item.path,
      name: item.name,
      path: item.path,
      parentId:
        item.path.split("/").length === 2
          ? "0"
          : dirs.find(
              (x) => x.path === item.path.split("/").slice(0, -1).join("/")
            )?.path,
      type: "FILE",
      depth: 0,
    };
    cache.set(file.id, file);
  });

  cache.forEach((value, key) => {
    if (value.parentId === "0") {
      if (value.type === "DIRECTORY") rootDir.dirs.push(value);
      else rootDir.files.push(value);
    } else {
      const parentDir = cache.get(value.parentId);
      if (value.type === "DIRECTORY") parentDir.dirs.push(value);
      else parentDir.files.push(value);
    }
  });

  getDepth(rootDir, 0);

  return rootDir;
}
