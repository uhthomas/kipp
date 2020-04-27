def tag(branch):
    if branch == "master":
        return "latest"
    return "local"

def tags():
    if "$(CI)" != "true":
        return []
    return [
        "$(GITHUB_SHA)",
        "latest-$(GITHUB_REF)",
    ]