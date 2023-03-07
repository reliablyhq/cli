IS_WINDOWS = "windows" in BUILD_TARGET_TRIPLE
IS_LINUX = "linux" in BUILD_TARGET_TRIPLE
IS_APPLE = "apple" in BUILD_TARGET_TRIPLE


def make_exe():
    dist = default_python_distribution(python_version="3.10")

    policy = dist.make_python_packaging_policy()
    policy.set_resource_handling_mode("files")
    policy.resources_location = "filesystem-relative:lib"
    policy.resources_location_fallback = "filesystem-relative:lib"

    policy.include_test = False
    policy.include_non_distribution_sources = True
    policy.include_distribution_resources = True
    policy.include_distribution_sources = True
    policy.allow_files = True
    policy.include_file_resources = True
    policy.file_scanner_emit_files = True

    python_config = dist.make_python_interpreter_config()
    python_config.module_search_paths = ["$ORIGIN/lib"]
    python_config.filesystem_importer = True
    python_config.oxidized_importer = False
    python_config.sys_frozen = True

    python_config.run_command = "from reliably_cli.__main__ import cli; cli()"

    exe = dist.to_python_executable(
        name="reliably",
        packaging_policy=policy,
        config=python_config,
    )
    
    exe.windows_runtime_dlls_mode = "always"
    exe.windows_subsystem = "console"

    if not IS_LINUX:
        # pip download seems preferred over pip install in cross compilation
        # scenarios https://github.com/indygreg/PyOxidizer/issues/566#issuecomment-1146851507
        exe.add_python_resources(exe.pip_download(["reliably-cli"]))
    else:
        exe.add_python_resources(
            exe.pip_install(["reliably-cli"], {"PIP_NO_BINARY": "pydantic"})
        )

    return exe

def make_embedded_resources(exe):
    return exe.to_embedded_resources()

def make_install(exe):
    # Create an object that represents our installed application file layout.
    files = FileManifest()

    # Add the generated executable to our install layout in the root directory.
    files.add_python_resource(".", exe)

    return files

def make_msi(exe):
    return exe.to_wix_msi_builder(
        "reliably",
        "Reliably CLI",
        "0.1.1",
        "ChaosIQ Ltd"
    )

# Tell PyOxidizer about the build targets defined above.
register_target("exe", make_exe)
register_target("resources", make_embedded_resources, depends=["exe"], default_build_script=True)
register_target("install", make_install, depends=["exe"], default=True)
register_target("msi", make_msi, depends=["exe"])

# Resolve whatever targets the invoker of this configuration file is requesting
# be resolved.
resolve_targets()
