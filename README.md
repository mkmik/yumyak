Clone and (go) run in the current dir.

From another shell run a centos image and run:

    echo "$YOURIP mirror1" >>/etc/hosts
    yum-config-manager --add-repo http://$YOURIP:8080/yak.repo
    yum search yakshaver

