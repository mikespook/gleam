import urllib

def main(root, args):
    import os
    import znode
    import urllib
    if len(args) < 2:
        znode.error("not enough args: %s" % str(args))
        return
    head, tail = os.path.split(args[1])
    r = '%s/%s' % (root, tail)

    try:
        urllib.urlretrieve(args[0], filename=r)
        znode.debug("saved %s to %s" % (args[0], r))
    except Exception, e:
        znode.error("error downloading %s to %s: %s" % (args[0], r, e))


main(_ROOT_, _ARGS_)
