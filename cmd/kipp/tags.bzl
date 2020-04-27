def tags():
    if "$(CI)" != "true":
        return []
    return [
        "$(GITHUB_SHA)",
        "latest-$(GITHUB_REF)",
    ]