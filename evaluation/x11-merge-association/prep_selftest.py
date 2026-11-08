#!/usr/bin/env python3
"""Synthetic-only preparation tests. Running these never opens a candidate clone."""
import argparse
import copy
import json
from pathlib import Path
import tempfile
from types import SimpleNamespace
import unittest
from unittest.mock import patch
import inventory

import prepare_exports as prep
import assemble_preparation
import pin_tool_closure as closure
import stages


def write_json(path, value):
    path.write_text(json.dumps(value, sort_keys=True, indent=2) + '\n')


def blob_id(data):
    return prep.hashlib.sha1(b'blob ' + str(len(data)).encode() + b'\0' + data).hexdigest()


def synthetic_artifact_chain(root, config, inputs, authority, ownership_failure=None, build_failure=None):
    """Produce real sealed operations on a synthetic repository and Go package."""
    tools = {key: prep.pin(value) for key, value in config['export_preparation']['tools'].items()}
    project = root / 'synthetic-go'; project.mkdir()
    (project / 'go.mod').write_text('module example.invalid/preparation-fixture\n\ngo 1.24\n')
    (project / 'go.sum').write_text('')
    (project / 'main.go').write_text('package main\nfunc main() {}\n')
    recipe = prep.load(prep.pin(config['export_preparation']['frontend_build_recipe']))
    binary, buildinfo = project/'built/frontend', project/'built/go-version-m.txt'
    recipe.update(cwd=str(project), argv=[str(tools['go']), 'build', '-pgo=off', '-trimpath',
                  '-buildvcs=false', '-o', str(binary), '.'],
                  version_m_argv=[str(tools['go']), 'version', '-m', str(binary)],
                  required_buildinfo_path=str(buildinfo))
    env = dict(recipe['environment'])
    for key in ('HOME', 'GOCACHE', 'GOTMPDIR'):
        location = project/key.lower(); location.mkdir()
        env[key] = str(location)
    metadata = closure.go_metadata(recipe, env, lambda argv, cwd, env: prep.run(tools, env, argv, cwd))
    closure_path = project/'closure.json'; write_json(closure_path, metadata)
    recipe['closure'] = prep.file_pin(closure_path)
    recipe_path = project/'recipe.json'; write_json(recipe_path, recipe)
    config['export_preparation']['frontend_build_recipe'] = prep.file_pin(recipe_path)
    write_json(inputs, config)
    signed = prep.load(authority)
    signed['files'] = [entry for entry in signed['files'] if Path(entry['path']) != inputs]
    signed['files'].extend([prep.file_pin(inputs), prep.file_pin(recipe_path), prep.file_pin(closure_path),
                            *prep.script_pins()])
    write_json(authority, signed)
    budget = prep.canonical_budget(inputs)
    prep.start_budget(budget, inputs, authority)
    seals = []
    def execute(operation, repository=None):
        output, seal = prep.canonical_paths(inputs, operation, repository)
        args = SimpleNamespace(inputs=str(inputs), authorization=str(authority), budget=str(budget), operation=operation,
            repository=repository, scratch=str(root/'scratch'),
            output=str(output), seal=str(seal))
        if operation == 'build-frontend' and build_failure:
            with patch.object(prep, 'run', build_failure):
                record = prep.execute(args)
        elif operation == 'ownership' and ownership_failure:
            with patch.object(prep, 'run', ownership_failure):
                record = prep.execute(args)
        else:
            record = prep.execute(args)
        prep.verify_preparation_seal(Path(args.seal), config, inputs, authority)
        seals.append(Path(args.seal))
        return record
    execute('build-frontend')
    execute('exposures')
    for item in config['repositories']:
        execute('ownership', item['repository'])
    artifacts = assemble_preparation.assemble(inputs, authority)
    return artifacts, [prep.file_pin(path) for path in seals]


class PreparationTests(unittest.TestCase):
    inputs_path = prep.HERE / 'inventory-inputs.json'
    scratch_root = None

    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory(prefix='x11-prep-synthetic-', dir=self.scratch_root)
        self.addCleanup(self.temporary.cleanup)
        self.root = Path(self.temporary.name).resolve()
        self.config = prep.load(self.inputs_path)
        self.tools = {name: prep.pin(pin) for name, pin in self.config['export_preparation']['tools'].items()}
        home, binaries = self.root / 'home', self.root / 'bin'
        home.mkdir(); binaries.mkdir()
        (binaries / 'git').symlink_to(self.tools['git'])
        self.env = prep.environment(home, binaries)
        prep.TRACE.clear()

    def command(self, *args, cwd=None, data=None):
        return prep.run(self.tools, self.env, args, cwd or self.root, data)

    def repository(self, unsafe=False):
        repo = self.root / 'synthetic-repo'
        repo.mkdir()
        self.command(self.tools['git'], 'init', '-q', str(repo))
        (repo / 'docs').mkdir()
        (repo / 'docs' / 'guide.txt').write_bytes(b'literal $Format:%H$\n')
        (repo / 'README.md').write_bytes(b'# Synthetic project\n\n## Feature\nDocumentation fixture.\n')
        (repo / '.gitattributes').write_text('/docs export-ignore export-subst\n/docs/guide.txt export-subst\n/README.md export-ignore export-subst\n')
        if unsafe:
            (repo / 'unsafe-link').symlink_to('/outside-synthetic-root')
        self.command(self.tools['git'], '-C', str(repo), 'add', '.')
        self.command(self.tools['git'], '-C', str(repo), '-c', 'user.name=Synthetic',
                     '-c', 'user.email=synthetic@example.invalid', 'commit', '-qm', 'synthetic fixture')
        commit = self.command(self.tools['git'], '-C', str(repo), 'rev-parse', 'HEAD').decode().strip()
        return repo, commit

    def authority(self, repo, commit):
        config = copy.deepcopy(self.config)
        config['repositories'] = [{'repository': 'synthetic/preparation', 'path': str(repo),
            'default_branch': 'master', 'resolved_remote_tip': commit, 'tip': commit,
            'Cold': commit, 'C0': commit, 'canonical_clone_url': 'https://example.invalid/synthetic.git'}]
        inputs = self.root / 'synthetic-inputs.json'
        write_json(inputs, config)
        review = self.root / 'synthetic-review'
        review.write_text('Synthetic fixture authority only; no candidate authority.')
        pins = [prep.file_pin(stages.ROOT / name) for name in stages.REVIEWED]
        pins += [prep.file_pin(inputs), *config['recipe_sources'],
                 *config['export_preparation']['tools'].values(),
                 config['export_preparation']['clawpatch_tree'],
                 config['export_preparation']['frontend_build_recipe'],
                 config['export_preparation']['python_runtime'],
                 config['export_preparation']['exposure_extraction']['seal'],
                 *config['export_preparation']['exposure_extraction']['files'],
                 *prep.script_pins()]
        build = prep.load(prep.pin(config['export_preparation']['frontend_build_recipe']))
        pins.append(build['closure'])
        auth = self.root / 'synthetic-authority.json'
        write_json(auth, {'scope': 'export-preparation', 'verdict': 'ZERO ISSUES',
                          'review': prep.file_pin(review), 'files': pins})
        return config, inputs, auth

    def budget(self, inputs, authority):
        path = prep.canonical_budget(inputs)
        if not path.exists():
            prep.start_budget(path, inputs, authority)
        return str(path)

    def args(self, inputs, auth):
        output, seal = prep.canonical_paths(inputs, 'ownership', 'synthetic/preparation')
        return SimpleNamespace(inputs=str(inputs), authorization=str(auth), budget=self.budget(inputs, auth), operation='ownership',
            repository='synthetic/preparation', scratch=str(self.root / 'scratch'),
            output=str(output), seal=str(seal))

    def test_exact_blobs_ignore_export_attributes(self):
        repo, commit = self.repository()
        root = self.root / 'materialized'; root.mkdir()
        prep.materialize(self.tools, self.env, repo, commit, root)
        self.assertEqual((root / 'docs/guide.txt').read_bytes(), b'literal $Format:%H$\n')
        self.assertEqual((root / 'README.md').read_bytes(), (repo / 'README.md').read_bytes())
        self.assertEqual((root / '.gitattributes').read_bytes(), (repo / '.gitattributes').read_bytes())
        self.assertTrue(any('cat-file' in c.get('argv', []) for c in prep.TRACE))
        self.assertFalse(any('archive' in c.get('argv', []) or 'check-attr' in c.get('argv', []) for c in prep.TRACE))

    def test_scanner_failures_are_always_operational(self):
        for code, stderr in [(1, b'content-looking'), (-9, b''), (137, b''),
                             (1, b'EROFS'), (1, b'EDQUOT'), (1, b'ENOENT'), (1, b'ENOMEM')]:
            failure = prep.CommandFailure(SimpleNamespace(returncode=code, stdout=b'', stderr=stderr),
                                          ['node', 'scanner', 'map'])
            self.assertIsInstance(failure, prep.OperationalRefusal)
            self.assertNotIsInstance(failure, prep.ContentRefusal)

    def test_safety_refusals(self):
        cases = [('../escape', '100644', b'x'), ('/absolute', '100644', b'x'),
                 ('nested/.codegraph/secret', '100644', b'x'), ('module', '100664', b'x'),
                 ('link', '120000', b'/absolute'), ('link', '120000', b'../escape'),
                 ('link', '120000', b'.codegraph/secret'), ('link', '120000', b'\xff')]
        for index, (name, mode, data) in enumerate(cases):
            with self.subTest(name=name, mode=mode):
                target = self.root / str(index); target.mkdir()
                with self.assertRaises(prep.ContentRefusal):
                    prep.materialize_entries([(name, mode, '1'*40)], target, lambda oid: data)
        target = self.root / 'ancestor'; target.mkdir()
        with self.assertRaises(prep.ContentRefusal):
            prep.materialize_entries([('link', '120000', '1'*40),
                                      ('link/child', '100644', '2'*40)], target, lambda oid: b'.')
        target = self.root / 'chain'; target.mkdir()
        with self.assertRaises(prep.ContentRefusal):
            prep.materialize_entries([('a', '120000', '1'*40), ('b', '120000', '2'*40)],
                                     target, lambda oid: b'b' if oid == '1'*40 else b'a')

    def test_safe_symlink_and_executable_mode(self):
        target = self.root / 'safe'; target.mkdir()
        executable, link = b'#!/bin/sh\nexit 0\n', b'bin/run'
        blobs = {blob_id(executable): executable, blob_id(link): link}
        prep.materialize_entries([('bin/run', '100755', blob_id(executable)),
                                  ('run-link', '120000', blob_id(link))], target, blobs.__getitem__)
        self.assertEqual((target/'bin/run').stat().st_mode & 0o777, 0o755)
        self.assertTrue((target/'run-link').is_symlink())
        self.assertEqual((target/'run-link').read_bytes(), executable)
        dangling = self.root / 'dangling'; dangling.mkdir()
        target_bytes = b'missing'
        prep.materialize_entries([('link', '120000', blob_id(target_bytes))],
                                 dangling, lambda oid: target_bytes)
        self.assertTrue((dangling/'link').is_symlink())
        self.assertEqual(prep.os.readlink(dangling/'link'), 'missing')

    def test_materialized_selected_tree_mismatches(self):
        blobs = {blob_id(b'original'): b'original', blob_id(b'data'): b'data'}
        entries = [('data', '100644', blob_id(b'original')),
                   ('link', '120000', blob_id(b'data'))]
        for mismatch in ('bytes', 'mode', 'extra', 'symlink'):
            with self.subTest(mismatch=mismatch):
                target = self.root / mismatch; target.mkdir()
                prep.materialize_entries(entries, target, blobs.__getitem__)
                if mismatch == 'bytes':
                    (target/'data').write_bytes(b'changed')
                elif mismatch == 'mode':
                    (target/'data').chmod(0o600)
                elif mismatch == 'extra':
                    (target/'extra').write_bytes(b'extra')
                else:
                    (target/'link').unlink()
                    (target/'link').symlink_to('other')
                with self.assertRaises(ValueError) as raised:
                    prep.verify_materialized(entries, target)
                self.assertNotIsInstance(raised.exception, prep.ContentRefusal)

    def test_unicode_alias_safety(self):
        target = self.root / 'unicode'; target.mkdir()
        blobs = {blob_id(b'A'): b'A', blob_id(b'B'): b'B'}
        entries = [('name-\u00e9', '100644', blob_id(b'A')),
                   ('name-e\u0301', '100644', blob_id(b'B'))]
        with self.assertRaises(prep.ContentRefusal):
            prep.materialize_entries(entries, target, blobs.__getitem__)

    def test_case_safety(self):
        target = self.root / 'case'; target.mkdir()
        blobs = {blob_id(b'A'): b'A', blob_id(b'B'): b'B'}
        entries = [('Case', '100644', blob_id(b'A')), ('case', '100644', blob_id(b'B'))]
        with self.assertRaises(prep.ContentRefusal):
            prep.materialize_entries(entries, target, blobs.__getitem__)

    def test_authorization_scope_and_unsigned_implementation(self):
        repo, commit = self.repository()
        _, inputs, auth = self.authority(repo, commit)
        record = prep.load(auth)
        record['scope'] = 'inventory'; write_json(auth, record)
        with self.assertRaises(ValueError):
            prep.authorise(inputs, auth)
        record['scope'] = 'export-preparation'
        record['files'] = [p for p in record['files'] if Path(p['path']) != Path(prep.__file__).resolve()]
        write_json(auth, record)
        with self.assertRaises(ValueError):
            prep.authorise(inputs, auth)

    def test_scanner_tree_drift_before_command(self):
        expected = prep.load(prep.pin(self.config['export_preparation']['clawpatch_tree']))
        expected['files'][str(self.tools['clawpatch'])] = '0'*64
        drift = self.root / 'drift.json'; write_json(drift, expected)
        config = copy.deepcopy(self.config['export_preparation'])
        config['clawpatch_tree'] = prep.file_pin(drift)
        before = len(prep.TRACE)
        with self.assertRaisesRegex(ValueError, 'whole scanner install changed'):
            prep.verify_scanner(config, self.tools)
        self.assertEqual(len(prep.TRACE), before)

    def test_assembler_derives_manifest_from_verified_seals(self):
        repo, commit = self.repository()
        config, inputs, auth = self.authority(repo, commit)
        artifacts, _ = synthetic_artifact_chain(self.root, config, inputs, auth)
        seals = [Path(union['seal']['path']) for union in artifacts['ownership'].values()]
        seals += [Path(entry['path']) for entry in artifacts['preparation_seals']]
        assembled = assemble_preparation.assemble(inputs, auth)
        self.assertEqual(assembled, artifacts)
        original = seals[0].read_bytes()
        seals[0].unlink()
        with self.assertRaises(FileNotFoundError):
            assemble_preparation.assemble(inputs, auth)
        seals[0].write_bytes(original)
        with self.assertRaises(TypeError):
            assemble_preparation.assemble(inputs, auth, seals)
        with self.assertRaises(TypeError):
            assemble_preparation.assemble(inputs, auth, [*seals, seals[0]])
        record = prep.load(seals[0])
        record['inputs']['sha256'] = '0'*64
        write_json(seals[0], record)
        with self.assertRaisesRegex(ValueError, 'authority/input mismatch'):
            assemble_preparation.assemble(inputs, auth)

    def test_ownership_real_scanner_create_once_and_seals(self):
        repo, commit = self.repository()
        config, inputs, auth = self.authority(repo, commit)
        args = self.args(inputs, auth)
        record = prep.execute(args)
        self.assertEqual(record['status'], 'completed', record['refusal'])
        verified = prep.verify_preparation_seal(Path(args.seal), config, inputs, auth)
        self.assertEqual(verified['revision_count'], 1)
        self.assertTrue(any('heuristic' in command.get('argv', []) for command in record['commands']))
        with self.assertRaises(ValueError):
            prep.execute(args)
        original = Path(args.output).read_bytes()
        Path(args.output).write_bytes(original + b' ')
        with self.assertRaises(ValueError):
            prep.verify_preparation_seal(Path(args.seal), config, inputs, auth)

    def test_noncanonical_operation_paths_refuse_before_commands(self):
        repo, commit = self.repository()
        _, inputs, auth = self.authority(repo, commit)
        args = self.args(inputs, auth)
        before = len(prep.TRACE)
        for field in ('output', 'seal'):
            alternate = SimpleNamespace(**vars(args))
            setattr(alternate, field, str(self.root / ('alternate-' + field + '.json')))
            with self.subTest(field=field), self.assertRaisesRegex(ValueError, 'canonical paths'):
                prep.execute(alternate)
            self.assertFalse(Path(alternate.output).exists())
            self.assertFalse(Path(alternate.seal).exists())
        self.assertEqual(len(prep.TRACE), before)

    def test_content_and_operational_refusal_seals(self):
        repo, commit = self.repository(unsafe=True)
        config, inputs, auth = self.authority(repo, commit)
        args = self.args(inputs, auth)
        record = prep.execute(args)
        self.assertEqual(record['status'], 'refused')
        self.assertEqual(record['refusal']['class'], 'content')
        self.assertFalse(Path(args.output).exists())
        prep.verify_preparation_seal(Path(args.seal), config, inputs, auth)
        config['repositories'][0]['path'] = str(self.root / 'absent-synthetic-repo')
        old_inputs = inputs
        folder = self.root/'operational'; folder.mkdir()
        inputs = folder/'inputs.json'
        write_json(inputs, config)
        authority = prep.load(auth)
        authority['files'] = [prep.file_pin(inputs) if Path(p['path']) == old_inputs else p for p in authority['files']]
        write_json(auth, authority)
        args = self.args(inputs, auth)
        record = prep.execute(args)
        self.assertEqual(record['refusal']['class'], 'operational')
        prep.verify_preparation_seal(Path(args.seal), config, inputs, auth)
        with self.assertRaises(FileExistsError):
            prep.execute(args)

    def test_build_added_source_changes_selected_closure(self):
        # A dependency-free synthetic Go package; closure drift must refuse before build.
        project = self.root / 'go-project'; project.mkdir()
        (project / 'go.mod').write_text('module example.invalid/synthetic\n\ngo 1.24\n')
        (project / 'go.sum').write_text('')
        (project / 'main.go').write_text('package main\nfunc main() {}\n')
        recipe = prep.load(prep.pin(self.config['export_preparation']['frontend_build_recipe']))
        recipe['cwd'] = str(project)
        recipe['argv'] = [str(self.tools['go']), 'build', '-pgo=off', '-o', str(project/'outputs/never-built'), '.']
        recipe['version_m_argv'] = [str(self.tools['go']), 'version', '-m', str(project/'outputs/never-built')]
        recipe['required_buildinfo_path'] = str(project/'outputs/never-written-buildinfo')
        environment = dict(recipe['environment'])
        for key in ('HOME', 'GOCACHE', 'GOTMPDIR'):
            directory = self.root / ('private-' + key); directory.mkdir()
            environment[key] = str(directory)
        runner = lambda argv, cwd, env: prep.run(self.tools, env, argv, cwd)
        expected = closure.go_metadata(recipe, environment, runner)
        closure_path = project/'closure.json'; write_json(closure_path, expected)
        recipe['closure'] = prep.file_pin(closure_path)
        recipe_path = project/'recipe.json'; write_json(recipe_path, recipe)
        config = copy.deepcopy(self.config['export_preparation'])
        config['frontend_build_recipe'] = prep.file_pin(recipe_path)
        (project / 'added.go').write_text('package main\nvar Added = true\n')
        with self.assertRaisesRegex(ValueError, 'selected package/file-set'):
            prep.build_frontend(config, self.tools, lambda path: None)
        self.assertFalse((project/'never-built').exists())
        self.assertFalse(any('build' in command.get('argv', []) for command in prep.TRACE))

    def test_build_refuses_existing_artifact(self):
        # Existing artifact refusal occurs before any build command.
        recipe = prep.load(prep.pin(self.config['export_preparation']['frontend_build_recipe']))
        binary = self.root / 'existing-binary'; binary.touch()
        recipe['argv'][recipe['argv'].index('-o') + 1] = str(binary)
        recipe['version_m_argv'][-1] = str(binary)
        recipe['required_buildinfo_path'] = str(binary.parent/'info')
        path = self.root / 'build-recipe.json'; write_json(path, recipe)
        config = copy.deepcopy(self.config['export_preparation'])
        config['frontend_build_recipe'] = prep.file_pin(path)
        before = len(prep.TRACE)
        with self.assertRaisesRegex(ValueError, 'create-once'):
            prep.build_frontend(config, self.tools, lambda path: None)
        self.assertEqual(len(prep.TRACE), before)


    def test_gitlinks_and_excluded_codegraph_counts(self):
        repo, commit = self.repository()
        # Index-only fixture: never create or modify a .codegraph directory.
        oid = self.command(self.tools['git'], '-C', str(repo), 'hash-object', '-w', '--stdin', data=b'excluded').decode().strip()
        self.command(self.tools['git'], '-C', str(repo), 'update-index', '--add', '--cacheinfo',
                     '100644,'+oid+',.codegraph/committed-record')
        self.command(self.tools['git'], '-C', str(repo), 'update-index', '--add', '--cacheinfo',
                     '160000,'+commit+',module')
        tree = self.command(self.tools['git'], '-C', str(repo), 'write-tree').decode().strip()
        revision = self.command(self.tools['git'], '-C', str(repo), '-c', 'user.name=Synthetic',
            '-c', 'user.email=synthetic@example.invalid', 'commit-tree', tree, '-p', commit, data=b'synthetic gitlink').decode().strip()
        root = self.root/'gitlink-tree'; root.mkdir()
        metadata = prep.materialize(self.tools, self.env, repo, revision, root)
        self.assertEqual(metadata['gitlinks'], 1)
        self.assertEqual(metadata['excluded_codegraph_entries'], 1)
        self.assertTrue((root/'module').is_dir())
        self.assertEqual(list((root/'module').iterdir()), [])
        self.assertFalse((root/'.codegraph').exists())

    def test_git_structural_operational_prerequisites(self):
        repo, commit = self.repository()
        common = repo/'.git'
        for path, data in ((common/'info/grafts', b''), (common/'shallow', (commit+'\n').encode()),
                           (common/'refs/replace'/commit, (commit+'\n').encode()),
                           (common/'objects/pack/synthetic.promisor', b'')):
            path.parent.mkdir(parents=True, exist_ok=True)
            path.write_bytes(data)
            with self.subTest(path=path), self.assertRaises(ValueError) as raised:
                prep.repository_prerequisites(self.tools, self.env, repo)
            self.assertNotIsInstance(raised.exception, prep.ContentRefusal)
            path.unlink()
        self.command(self.tools['git'], '-C', str(repo), 'config', 'remote.origin.promisor', 'true')
        with self.assertRaises(ValueError):
            prep.repository_prerequisites(self.tools, self.env, repo)
        self.assertEqual(self.env['GIT_NO_REPLACE_OBJECTS'], '1')

    def test_scanner_retry_sealing(self):
        repo, commit = self.repository()
        config, inputs, authority = self.authority(repo, commit)
        original_run = prep.run
        failures = []
        def once(tools, env, argv, cwd, *args, **kwargs):
            if list(map(str, argv))[-3:] == ['map', '--source', 'heuristic'] and not failures:
                failures.append(True)
                return original_run(tools, env, [tools['node'], '-e', 'process.exit(19)'], cwd)
            return original_run(tools, env, argv, cwd, *args, **kwargs)
        args = self.args(inputs, authority)
        with patch.object(prep, 'run', once):
            record = prep.execute(args)
        self.assertEqual(record['status'], 'completed', record['refusal'])
        self.assertEqual(len(record['attempts']), 2)
        prep.verify_preparation_seal(Path(args.seal), config, inputs, authority)

    def test_memoryerror_sealing(self):
        repo, commit = self.repository()
        config, inputs, authority = self.authority(repo, commit)
        args = self.args(inputs, authority)
        with patch.object(prep, 'heuristic_paths', side_effect=MemoryError):
            record = prep.execute(args)
        self.assertEqual(record['refusal']['code'], 'PreparationInfeasibleOperational')
        self.assertEqual(len(record['attempts']), 3)
        self.assertTrue(all(attempt['code'] == 'MemoryError' for attempt in record['attempts']))
        prep.verify_preparation_seal(Path(args.seal), config, inputs, authority)
        with self.assertRaises(FileExistsError):
            prep.execute(args)

    def test_produced_refusals_through_inventory(self):
        for operational in (False, True):
            with self.subTest(operational=operational):
                fixture = PreparationTests()
                fixture.inputs_path = self.inputs_path
                fixture.setUp()
                self.addCleanup(fixture.doCleanups)
                repo, commit = fixture.repository(unsafe=not operational)
                config, inputs, authority = fixture.authority(repo, commit)
                original_run = prep.run
                def fail_scanner(tools, env, argv, cwd, *args, **kwargs):
                    if list(map(str, argv))[-3:] == ['map', '--source', 'heuristic']:
                        return original_run(tools, env, [tools['node'], '-e', 'process.exit(17)'], cwd)
                    return original_run(tools, env, argv, cwd, *args, **kwargs)
                artifacts, _ = synthetic_artifact_chain(fixture.root, config, inputs, authority,
                                                        fail_scanner if operational else None)
                manifest = fixture.root/'artifacts.json'; write_json(manifest, artifacts)
                inventory.verify_artifact_bindings(artifacts, inputs, fixture.root)
                result = inventory.inventory(inputs, manifest)
                self.assertEqual(result['selection']['verdict'],
                    'not_testable_operational' if operational else 'not_testable_no_substitution')
                refusal = result['repositories'][0]['preparation_refusal']
                self.assertEqual(refusal['class'], 'operational' if operational else 'content')
                if operational:
                    self.assertEqual(refusal['code'], 'PreparationInfeasibleOperational')

    def test_completed_seal_rejects_empty_or_inconsistent_trace(self):
        repo, commit = self.repository()
        config, inputs, authority = self.authority(repo, commit)
        args = self.args(inputs, authority)
        record = prep.execute(args)
        self.assertEqual(record['status'], 'completed', record['refusal'])
        for mutation in ('empty', 'preflight-missing', 'preflight-input', 'revision', 'count', 'digest'):
            broken = copy.deepcopy(record)
            if mutation == 'empty':
                broken['commands'] = []
            elif mutation == 'preflight-missing':
                broken['commands'] = [entry for entry in broken['commands'] if entry.get('step') != 'preflight']
            elif mutation == 'preflight-input':
                step = next(entry for entry in broken['commands'] if entry.get('step') == 'preflight')
                step['inputs']['sha256'] = '0'*64
            else:
                step = next(entry for entry in broken['commands'] if entry.get('step') == 'materialized')
                if mutation == 'revision':
                    step['revision'] = '0'*40
                elif mutation == 'count':
                    step['count'] += 1
                else:
                    step['oid_list_sha256'] = '0'*64
            write_json(Path(args.seal), broken)
            with self.subTest(mutation=mutation), self.assertRaises(ValueError):
                prep.verify_preparation_seal(Path(args.seal), config, inputs, authority)


    def test_timeout_and_interrupt_are_operational(self):
        original_run = prep.run
        for kind in ('timeout', 'interrupt'):
            fixture = PreparationTests()
            fixture.inputs_path = self.inputs_path
            fixture.setUp()
            self.addCleanup(fixture.doCleanups)
            repo, commit = fixture.repository()
            config, inputs, authority = fixture.authority(repo, commit)
            def fail(tools, env, argv, cwd, *extra, **kwargs):
                if list(map(str, argv))[-3:] == ['map', '--source', 'heuristic']:
                    if kind == 'interrupt':
                        raise KeyboardInterrupt()
                    with patch.object(prep, 'SUBPROCESS_SECONDS', 0.05):
                        return original_run(tools, env, [tools['node'], '-e',
                            "require('child_process').spawn(process.execPath,['-e','setInterval(()=>{},1000)']);setInterval(()=>{},1000)"], cwd)
                return original_run(tools, env, argv, cwd, *extra, **kwargs)
            args = fixture.args(inputs, authority)
            with patch.object(prep, 'run', fail):
                record = prep.execute(args)
            self.assertEqual(record['refusal']['class'], 'operational')
            self.assertEqual(len(record['attempts']), 1 if kind == 'interrupt' else 3)
            self.assertEqual(record['refusal']['code'], 'PreparationInterruptedOperational'
                             if kind == 'interrupt' else 'PreparationInfeasibleOperational')
            prep.verify_preparation_seal(Path(args.seal), config, inputs, authority)
            if kind == 'interrupt':
                with self.assertRaisesRegex(ValueError, 'study-wide interruption/deadline'):
                    assemble_preparation.assemble(inputs, authority)
                other = SimpleNamespace(**vars(args))
                other.operation, other.repository = 'exposures', None
                output, seal = prep.canonical_paths(inputs, other.operation)
                other.output, other.seal = str(output), str(seal)
                with self.assertRaisesRegex(ValueError, 'study-wide interruption/deadline'):
                    prep.execute(other)
                self.assertFalse(seal.exists())
                # The obsolete list interface cannot omit the interrupt.
                with self.assertRaises(TypeError):
                    assemble_preparation.assemble(inputs, authority, [])

    def test_persistent_budget_expiry_is_global_even_before_commands(self):
        repo, commit = self.repository()
        config, inputs, authority = self.authority(repo, commit)
        args = self.args(inputs, authority)
        # Synthetic clock fixture, not a production budget reset.
        budget = prep.load(Path(args.budget))
        started = prep.datetime.now(prep.timezone.utc) - prep.timedelta(days=43)
        budget['started_utc'] = started.isoformat()
        budget['deadline_utc'] = (started + prep.timedelta(days=42)).isoformat()
        write_json(Path(args.budget), budget)
        record = prep.execute(args)
        self.assertEqual(record['refusal']['code'], 'PreparationBudgetExhausted')
        self.assertEqual(len(record['attempts']), 1)
        self.assertFalse(any('argv' in entry for entry in record['commands']))
        prep.verify_preparation_seal(Path(args.seal), config, inputs, authority)
        with self.assertRaisesRegex(ValueError, 'study-wide interruption/deadline'):
            assemble_preparation.assemble(inputs, authority)
        other = SimpleNamespace(**vars(args))
        other.operation, other.repository = 'exposures', None
        output, seal = prep.canonical_paths(inputs, other.operation)
        other.output, other.seal = str(output), str(seal)
        with self.assertRaisesRegex(ValueError, 'study-wide interruption/deadline'):
            prep.execute(other)
        self.assertFalse(seal.exists())

    def test_forged_content_refusal_has_no_git_evidence(self):
        repo, commit = self.repository(unsafe=True)
        config, inputs, authority = self.authority(repo, commit)
        args = self.args(inputs, authority)
        record = prep.execute(args)
        self.assertEqual(record['refusal']['class'], 'content')
        self.assertEqual(record['refusal']['revision'], commit)
        self.assertEqual(record['refusal']['failed_step'], 'cat-file')
        prep.verify_preparation_seal(Path(args.seal), config, inputs, authority)
        for mutation in ('preflight-only', 'no-batch', 'wrong-revision', 'wrong-step'):
            broken = copy.deepcopy(record)
            if mutation == 'preflight-only':
                broken['commands'] = [c for c in broken['commands'] if c.get('step') == 'preflight']
            elif mutation == 'no-batch':
                broken['commands'] = [c for c in broken['commands'] if c.get('argv', [])[-2:] != ['cat-file', '--batch']]
            elif mutation == 'wrong-revision':
                broken['refusal']['revision'] = '0'*40
            else:
                broken['refusal']['failed_step'] = 'invented'
            write_json(Path(args.seal), broken)
            with self.subTest(mutation=mutation), self.assertRaises(ValueError):
                prep.verify_preparation_seal(Path(args.seal), config, inputs, authority)

    def test_budget_path_cannot_reset_and_sixth_operation_refused(self):
        repo, commit = self.repository()
        config, inputs, authority = self.authority(repo, commit)
        args = self.args(inputs, authority)
        with self.assertRaises(FileExistsError):
            prep.start_budget(Path(args.budget), inputs, authority)
        alternate = self.root/'alternate-budget.json'
        with self.assertRaises(ValueError):
            prep.start_budget(alternate, inputs, authority)
        locks = []
        try:
            for number in range(5):
                slot = Path(args.budget + '.slot-' + str(number)).open('a')
                locks.append(slot)
                prep.fcntl.flock(slot, prep.fcntl.LOCK_EX | prep.fcntl.LOCK_NB)
            with self.assertRaisesRegex(prep.OperationalRefusal, 'five concurrent'):
                prep.execute(args)
        finally:
            for slot in locks:
                slot.close()

    def test_non_utf8_committed_path_is_structural_content(self):
        repo, _ = self.repository()
        oid = self.command(self.tools['git'], '-C', str(repo), 'hash-object', '-w', '--stdin', data=b'bytes').decode().strip()
        tree = self.command(self.tools['git'], '-C', str(repo), 'mktree', '-z',
                            data=b'100644 blob '+oid.encode()+b'\tbad-\xff\0').decode().strip()
        root = self.root/'non-utf8'; root.mkdir()
        with self.assertRaisesRegex(prep.ContentRefusal, 'non-UTF-8 committed path'):
            prep.materialize(self.tools, self.env, repo, tree, root)

    def test_buildinfo_failure_never_promotes_and_retry_recovers(self):
        repo, commit = self.repository()
        config, inputs, authority = self.authority(repo, commit)
        original_run = prep.run
        failures = []
        def fail_buildinfo_once(tools, env, argv, cwd, *args, **kwargs):
            if list(map(str, argv))[1:3] == ['version', '-m']:
                self.assertTrue(Path(argv[-1]).is_file())
                self.assertFalse((self.root/'synthetic-go/built').exists())
                if not failures:
                    failures.append(True)
                    return original_run(tools, env, [tools['node'], '-e', 'process.exit(23)'], cwd)
            return original_run(tools, env, argv, cwd, *args, **kwargs)
        artifacts, _ = synthetic_artifact_chain(self.root, config, inputs, authority,
                                               build_failure=fail_buildinfo_once)
        seal = next(prep.load(Path(entry['path'])) for entry in artifacts['preparation_seals']
                    if prep.load(Path(entry['path']))['operation'] == 'build-frontend')
        self.assertEqual(len(seal['attempts']), 2)
        self.assertTrue((self.root/'synthetic-go/built/frontend').is_file())
        self.assertTrue((self.root/'synthetic-go/built/go-version-m.txt').is_file())


    def test_revision_scope_excludes_unobserved_content_unit_only(self):
        # Unit-only index: scanner/materializer are replaced, so this output is
        # deliberately never passed to execute(), ASM or a production verifier.
        repo, cold = self.repository()
        commands = []
        for position in range(2002):
            message = ('synthetic %d' % position).encode()
            parent = cold if position == 0 else ':'+str(position)
            commands.append(b'commit refs/heads/synthetic\nmark :'+str(position+1).encode()
                +b'\ncommitter Synthetic <synthetic@example.invalid> 1 +0000\ndata '
                +str(len(message)).encode()+b'\n'+message+b'\nfrom '+parent.encode()+b'\n\n')
        self.command(self.tools['git'], '-C', str(repo), 'fast-import', '--quiet', data=b''.join(commands))
        chain = self.command(self.tools['git'], '-C', str(repo), 'rev-list', '--first-parent',
                             '--reverse', 'synthetic').decode().splitlines()
        config, inputs, authority = self.authority(repo, chain[-1])
        config['repositories'][0].update(Cold=cold, C0=chain[1])
        args = self.args(inputs, authority)
        # This unit calls ownership directly, bypassing execute's canonical reservation.
        Path(args.output).parent.mkdir(parents=True)
        seen = []
        original_run = prep.run
        def scan(tools, env, argv, cwd, *extra, **kwargs):
            if list(map(str, argv))[-1:] == ['init'] or list(map(str, argv))[-3:] == ['map', '--source', 'heuristic']:
                return b'{}'
            return original_run(tools, env, argv, cwd, *extra, **kwargs)
        def materialize(tools, env, repository, revision, root):
            if revision == chain[-1]:
                raise prep.ContentRefusal('unobserved tail content must not gate')
            seen.append(revision)
            return {}
        with patch.object(prep, 'run', scan), patch.object(prep, 'materialize', materialize), \
                patch.object(prep, 'heuristic_paths', return_value=[]):
            index = prep.ownership(args, config, config['export_preparation'], self.tools)
        self.assertEqual(len(index['revisions']), 2002)
        self.assertEqual(seen, [cold, *chain[1:2002]])
        self.assertNotIn(chain[-1], index['revisions'])


def run(inputs_path=None, scratch=None):
    if inputs_path:
        PreparationTests.inputs_path = Path(inputs_path)
    PreparationTests.scratch_root = Path(scratch) if scratch else None
    result = unittest.TextTestRunner(verbosity=2).run(unittest.defaultTestLoader.loadTestsFromTestCase(PreparationTests))
    if not result.wasSuccessful():
        raise AssertionError('synthetic preparation validation failed')


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--inputs', type=Path)
    parser.add_argument('--scratch', type=Path)
    args = parser.parse_args()
    run(args.inputs, args.scratch)
