IS_WINDOWS = "windows" in BUILD_TARGET_TRIPLE
IS_LINUX = "linux" in BUILD_TARGET_TRIPLE
IS_APPLE = "apple" in BUILD_TARGET_TRIPLE


def resource_callback(policy, resource):
    if type(resource) == "PythonModuleSource":
        resource.add_include = True


def make_exe():
    dist = default_python_distribution(python_version="3.10")

    policy = dist.make_python_packaging_policy()
    policy.register_resource_callback(resource_callback)

    policy.resources_location_fallback = "filesystem-relative:lib"
    policy.include_distribution_sources = False
    policy.include_non_distribution_sources = False
    policy.bytecode_optimize_level_zero = False
    policy.bytecode_optimize_level_one = False
    policy.bytecode_optimize_level_two = True

    python_config = dist.make_python_interpreter_config()
    python_config.module_search_paths = ["$ORIGIN/lib"]
    python_config.optimization_level = 2

    python_config.run_command = "from reliably_cli.__main__ import cli; cli()"

    exe = dist.to_python_executable(
        name="reliably",
        packaging_policy=policy,
        config=python_config,
    )
    
    exe.windows_runtime_dlls_mode = "always"
    exe.windows_subsystem = "console"

    exe.add_python_resources(exe.pip_download(["reliably-cli"]))

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
