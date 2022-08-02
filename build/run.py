import subprocess
import os

def subprocess_open(command):
    p = subprocess.Popen(command, stdout=subprocess.PIPE, stderr=subprocess.PIPE, universal_newlines=True)
    (stdoutdata, stderrdata) = p.communicate()
    return stdoutdata, stderrdata

def go_build(package, output):
    # out, err = subprocess_open(['go', 'build', '-a', '-o', output, package])
    out, err = subprocess_open(['go', 'build', '-o', output, package])
    if out != "" or err != "":
        return out, err
    return "", ""

def docker_build(name, tag, registry):
    image = registry + name + ":" + tag
    out, err = subprocess_open(['docker', 'build', '-t', image, "." ])
    if out != "" or err != "":
        return out, err
    return "", ""

def docker_push(name, tag, registry):
    image = registry + name + ":" + tag
    out, err = subprocess_open(['docker', 'push', image])
    if out != "" or err != "":
        return out, err
    return "", ""

def main():
    PKG_NAME = 'github.com/tmax-cloud/virtualrouter'
    GO_BINARY_NAME = 'virtualrouter'
    # DOCKER_REGISTRY = '10.0.0.4:5000/'
    DOCKER_REGISTRY = 'registry.network-team.tmaxanc.com/cloud/'
    # DOCKER_REGISTRY = "172.23.3.100/cloud/"
    DOCKER_IMAGE_NAME = "virtualrouter"
    DOCKER_IMAGE_TAG = "v0.2.1-dev"

    out, err = go_build(package=PKG_NAME, output=GO_BINARY_NAME)
    if err != "" or out != "":
        print("Error: " + err + ", Out: " + out)
        return
    print("Go build done")

    # checkout = docker_image_check(DOCKER_IMAGE_NAME,DOCKER_IMAGE_TAG,DOCKER_REGISTRY)
    # # print("checkout Type: " + type(checkout))
    # if checkout == "" :
    #     print("There is no image")
    # else :
    #     print("deleting "+ checkout)
    #     out, err = subprocess_open(['docker', 'rmi', checkout])
    # checkout = "$(docker images " + DOCKER_REGISTRY + DOCKER_IMAGE_NAME + ":" + DOCKER_IMAGE_TAG + " -q)"
    # out, err = subprocess_open(['docker', 'rmi', checkout])

    out, err = docker_build(name=DOCKER_IMAGE_NAME,tag=DOCKER_IMAGE_TAG, registry=DOCKER_REGISTRY)
    if err != "":
        print("Error: " + err)
        return
    print(out)
    print("Docker build done")

    out, err = docker_push(name=DOCKER_IMAGE_NAME,tag=DOCKER_IMAGE_TAG, registry=DOCKER_REGISTRY)
    if err != "":
        print("Error: " + err )
        return
    print(out)
    print("Docker push done")

    # os.chdir("../")
    # currentPath = os.getcwd()
    # print(currentPath)

def docker_image_check(name, tag, registry):
    image = registry + name + ":" + tag
    out, err = subprocess_open(['docker', 'images', '-f', 'reference=' + image, '-q'])
    if err != "":
        print("image_check_err")
        return 
    return out


if __name__ == "__main__":
    main()