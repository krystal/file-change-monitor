# File Change Monitor

This is a simple tool which can run a command while simultaneously monitoring a list of files for changes. If any of those files change, the running command will be sent a TERM signal.

## Why would you want this?

If you have an application that loads a file into memory when it starts and never reloads it, you find you have to manually restart it when changing the the file. This will only work when combined with some orchestration that will restart the process when it dies and when the disappearance of the process isn't service affecting.

Our use is we have processes that we run on Kubernetes which mount Secrets containing certificates that are renewed by a seperate process. When these certificates are renewed, we need to restart our container.

There are other solutions there for this but they require deploying an additional controller to the cluster, they must be running at the time the change happens for it to be noticed and changes to multiple secrest at the same time can result in multiple restarts.

## Usage

If you control your image, you can just add the `file-change-monitor` binary to that image and run that (although you'd probably be best to just have your own app handle changes to the underlying files itself).

If you don't control the image or the application, you can use an init container to use the file change monitor. This is an example of how you might do this.

```
kind: Pod
apiVersion: core/v1
metadata:
  name: example
spec:
  initContainers:
    - name: file-change-monitor
      image: registry.katapult.dev/generic/file-change-monitor:latest
      command:
        - /bin/cp
        - /file-change-monitor
        - /export/file-change-monitor
      volumeMounts:
        - name: file-change-monitor
          exportPath: /export
  containers:
    - name: myapp
      image: example/app:latest
      command:
        - /fcm/file-change-monitor
        - /config/example.crt
        - --
        - /myapp
        - --cert=/config/example.crt
      volumeMounts:
        - name: config
          exportPath: /config
          readOnly: true
        - name: file-change-monitor
          exportPath: /fcm
          readOnly: true
  volumes:
    - name: file-change-monitor
      emptyDir: {}
    - name: config
      configMap:
        name: app-config
```

You'll see here that you can just wrap any command with `file-monitor` by providing a list of config files to monitor.

```
$ file-monitor path/to/file1 path/to/file2 -- /my-app --other-options
```

By default, it will sleep 1 minute between checking the files (can override with the `--sleep` option). When a file changes, it will send a TERM signal to the process, if that doesn't kill the process in 5 minutes (override with `--kill-after`), it will send a KILL signal.
