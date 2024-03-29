IS_WINDOWS = "windows" in BUILD_TARGET_TRIPLE
RELIABLY_VERSION = VARS.get("RELIABLY_VERSION")


def make_exe():
    dist = default_python_distribution(python_version="3.10")

    policy = dist.make_python_packaging_policy()

    policy.allow_in_memory_shared_library_loading = True
    policy.bytecode_optimize_level_one = True
    policy.include_non_distribution_sources = False
    policy.include_test = False
    policy.resources_location = "in-memory"
    policy.resources_location_fallback = "filesystem-relative:lib"

    if IS_WINDOWS:
        policy.bytecode_optimize_level_one = True
        policy.extension_module_filter = "all"
        policy.include_file_resources = True
        policy.allow_files = True
        policy.file_scanner_emit_files = True
        policy.resources_location = "filesystem-relative:lib"

    python_config = dist.make_python_interpreter_config()
    python_config.module_search_paths = ["$ORIGIN", "$ORIGIN/lib"]
    python_config.run_command = "from reliably_cli.__main__ import cli; cli()"

    exe = dist.to_python_executable(
        name="reliably",
        packaging_policy=policy,
        config=python_config,
    )
    
    if IS_WINDOWS:
        exe.windows_runtime_dlls_mode = "always"
        exe.windows_subsystem = "console"
    
    exe.add_python_resources(exe.pip_install(["--prefer-binary", "--no-binary", "chaostoolkit-reliably", "-r", "requirements-generated.txt"]))
    exe.add_python_resources(exe.pip_install(["--prefer-binary", "--no-deps", "reliably-cli"]))

    return exe


def make_embedded_resources(exe):
    return exe.to_embedded_resources()

def make_install(exe):
    files = FileManifest()
    files.add_python_resource(".", exe)

    return files

def make_msi(exe):
    return exe.to_wix_msi_builder(
        "reliably",
        "Reliably",
        RELIABLY_VERSION,
        "ChaosIQ Ltd"
    )


def make_macos_app_bundle(exe):
    bundle = MacOsApplicationBundleBuilder("Reliably")
    bundle.set_info_plist_required_keys(
        display_name="Reliably",
        identifier="com.reliably.rbly",
        version=RELIABLY_VERSION,
        signature="rbly",
        executable="reliably",
    )

    universal = AppleUniversalBinary("reliably")
    universal.add_path("dist/x86_64-apple-darwin/reliably")

    m = FileManifest()
    m.add_file(universal.to_file_content())
    bundle.add_macos_manifest(m)

    return bundle


# Tell PyOxidizer about the build targets defined above.
register_target("exe", make_exe)
register_target("resources", make_embedded_resources, depends=["exe"], default_build_script=True)
register_target("install", make_install, depends=["exe"], default=True)
register_target("msi", make_msi, depends=["exe"])
register_target("macos", make_macos_app_bundle, depends=["exe"])

# Resolve whatever targets the invoker of this configuration file is requesting
# be resolved.
resolve_targets()
