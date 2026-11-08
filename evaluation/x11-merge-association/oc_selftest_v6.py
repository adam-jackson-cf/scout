"""Synthetic OC contract/refusal plants; standalone unittest entrypoint."""
import importlib.util
import json
from pathlib import Path
import tempfile
import unittest
from unittest.mock import Mock, patch

ROOT = Path(__file__).resolve().parent
spec = importlib.util.spec_from_file_location('x11a_oc', ROOT / 'oc_v5.py')
oc = importlib.util.module_from_spec(spec)
spec.loader.exec_module(oc)


def vectors(lengths=(1500, 1666, 2000)):
    value = {'segment_lengths': list(lengths),
             'counts': [[[0] * length for _ in range(2)] for length in lengths],
             'directory_cases': [[[] for _ in range(2)] for _ in range(3)]}
    value['counts'][0][0][0] = 1
    value['directory_cases'][0][0] = [{'position': 0, 'directory_hash': 'a' * 64}]
    return value


class OCContractTests(unittest.TestCase):
    def test_literal_plants(self):
        fixtures = json.loads((ROOT / 'oc-v5-fixtures.json').read_text())
        for row in fixtures['seed_table']:
            self.assertEqual(oc.scenario_index(*row['coordinates']), row['index'])
            self.assertEqual(oc.scenario_seed(row['index'], row['replicate'], row['stream']), row['seed'])
        for row in fixtures['count_plants']:
            self.assertEqual(oc.score_counts(**row['counts']), row['expected'])
        for row in fixtures['false_support_cases']:
            self.assertEqual(oc.false_support(row['truth'], row['supported']), row['expected_false'])
        for row in fixtures['segment_plants']:
            length = row['length']
            if row['eligible']:
                oc.validate_vectors(vectors((length,) * 3))
                self.assertEqual((length + 19) // 20, row['block_draws'])
                self.assertEqual(1 + (length - 1) % 20, row['final_block_positions'])
            else:
                with self.assertRaises(ValueError):
                    oc.validate_vectors(vectors((length,) * 3))

    def test_conditional_min_bound(self):
        # A valid directory component can veto a supporting commit component.
        self.assertEqual(oc.combined_bounds([.9] * 399, [.79] * 399), (.79, .9, True))
        # Negative requires BOTH upper bounds below .8, not the smallest upper.
        self.assertEqual(oc.combined_bounds([.7] * 399, [.81] * 399), (.7, .81, True))
        self.assertEqual(oc.combined_bounds([.9] * 399, [.9] * 378), (0., 1., False))
        self.assertEqual(oc.combined_bounds([.9] * 399, [.9] * 380), (.9, .9, True))
        self.assertEqual(oc.combined_bounds([], []), (0., 1., False))
        for invalid in (None, [None] * 399, [float('nan')] * 399,
                        [float('inf')] * 399, [1.1] * 399, [-.1] * 399):
            self.assertEqual(oc.combined_bounds([.9] * 399, invalid), (0., 1., False))

    def test_cli_refuses_non_timing_overrides(self):
        import sys
        original_flags = sys.flags
        class IsolatedFlags:
            isolated = 1
            def __getattr__(self, name):
                return getattr(original_flags, name)
        base = ['oc_v5.py', '--recipe', '/unread', '--build-record', '/unwritten',
                '--out', '/unwritten', '--mode', 'oc']
        for extra in (['--bootstrap', '10000'], ['--timing-replicates', '20'],
                      ['--directory-profile', 'singleton'], ['--segment-lengths', '1500,1666,2000']):
            with patch.object(sys, 'argv', base + extra), patch.object(
                    sys, 'flags', IsolatedFlags()), patch.object(
                    sys, 'dont_write_bytecode', True), patch.dict(
                    sys.modules, {'pin_tool_closure': Mock()}):
                with self.assertRaisesRegex(ValueError, 'forbids timing overrides'):
                    oc.main()

    def test_vector_refusals(self):
        oc.validate_vectors(vectors())
        for mutate in (
            lambda v: v.pop('segment_lengths'),
            lambda v: v.__setitem__('segment_lengths', None),
            lambda v: v['segment_lengths'].__setitem__(0, 1499),
            lambda v: v['segment_lengths'].__setitem__(0, 2001),
            lambda v: v['segment_lengths'].__setitem__(0, True),
            lambda v: v['segment_lengths'].__setitem__(0, None),
            lambda v: v['counts'][0].__setitem__(1, [0] * 1666),
            lambda v: v['counts'][0][0].__setitem__(0, None),
            lambda v: v['counts'][0].__setitem__(0, []),
            lambda v: v['directory_cases'][0][0][0].__setitem__('position', 1500),
            lambda v: v.pop('directory_cases'),
            lambda v: v['counts'][0][0].__setitem__(0, True),
            lambda v: v['counts'][0][0].__setitem__(0, 2),
            lambda v: v['directory_cases'][0][0][0].__setitem__('directory_hash', 'A' * 64),
            lambda v: v['directory_cases'][0][0][0].__setitem__('position', -1),
            lambda v: v.__setitem__('extra', []),
        ):
            value = vectors()
            mutate(value)
            with self.assertRaises(ValueError):
                oc.validate_vectors(value)
        for length in (1500, 1666, 2000):
            value = vectors((length,) * 3)
            oc.validate_vectors(value)

    def test_original_authority_output_and_recipe_bindings(self):
        with tempfile.TemporaryDirectory() as directory:
            base = Path(directory).resolve()
            def put(name, data):
                path = base / name
                path.write_text(json.dumps(data))
                return path
            def pin(path):
                return {'path': path.name, 'sha256': oc.sha(path)}
            recipe = put('recipe.json', {})
            inputs = put('inputs.json', {'recipe_sources': [pin(recipe)]})
            exports = put('exports.json', {'status': 'SEALED', 'source_inputs_sha256': oc.sha(inputs)})
            original = put('inventory-authority.json', {'scope': 'inventory'})
            output = put('inventory-output.json', {'oc_vectors': vectors()})
            vector = put('vectors.json', vectors())
            seal_data = {'inputs': [pin(inputs), pin(exports)], 'artifacts': [pin(output)]}
            seal = put('seal.json', seal_data)
            authority_data = {'reviewed_inventory': {'authorization': pin(original), 'seal': pin(seal)}}
            authority = put('protocol.json', authority_data)
            stages = Mock()
            stages.verify_authorization.return_value = authority_data
            stages.verify_seal.return_value = seal_data
            def bind():
                return oc.bind_oc_inputs(authority, inputs, output, vector, recipe, stages)
            result = bind()
            self.assertEqual(result['inventory_sha256'], oc.sha(output))
            stages.verify_authorization.assert_any_call(original, inputs, scope='inventory', artifacts=exports)
            stages.verify_seal.assert_called_with(seal, 'inventory', oc.sha(original), inputs)
            # A different but structurally valid vector cannot replace the sealed one.
            changed = vectors();changed['counts'][0][0][0] = 0;changed['directory_cases'][0][0] = []
            vector.write_text(json.dumps(changed))
            with self.assertRaisesRegex(ValueError, 'vectors differ'):
                bind()
            vector.write_text(json.dumps(vectors()))
            # A null original vector is refused even if a replacement is supplied.
            output.write_text(json.dumps({'oc_vectors': None}))
            seal_data['artifacts'] = [pin(output)]
            with self.assertRaisesRegex(ValueError, 'null'):
                bind()
            output.write_text(json.dumps({'oc_vectors': vectors()}));seal_data['artifacts'] = [pin(output)]
            authority_data.pop('reviewed_inventory')
            with self.assertRaisesRegex(ValueError, 'mandatory'):
                bind()
            authority_data['reviewed_inventory'] = {'authorization': pin(original), 'seal': pin(seal)}
            seal_data['artifacts'] = [pin(vector)]
            with self.assertRaisesRegex(ValueError, 'sole output'):
                bind()
            seal_data['artifacts'] = [pin(output)]
            stages.verify_authorization.side_effect = [authority_data, ValueError('original authority refused')]
            with self.assertRaisesRegex(ValueError, 'original authority refused'):
                bind()
            stages.verify_authorization.side_effect = None
            unpinned = put('unpinned.json', {})
            with self.assertRaisesRegex(ValueError, 'not source-pinned'):
                oc.bind_oc_inputs(authority, inputs, output, vector, unpinned, stages)

    def test_go_estimators_thresholds_and_vector_refusal(self):
        import subprocess
        recipe = json.loads((ROOT / 'tool-pins-v7/oc-build-recipe-v7.json').read_text())
        source = r'''
package main
import ("math"; "strings"; "testing")
func filled(x float64,n int)[]float64{v:=make([]float64,n);for i:=range v{v[i]=x};return v}
func validVector(length int)Vector{
 v:=Vector{SegmentLengths:[]int{length,length,length},Counts:make([][][]int,3),DirectoryCases:make([][][]DirectoryCase,3)}
 for repo:=range 3{v.Counts[repo]=make([][]int,2);v.DirectoryCases[repo]=make([][]DirectoryCase,2);for age:=range 2{v.Counts[repo][age]=make([]int,length);v.DirectoryCases[repo][age]=[]DirectoryCase{}}};return v
}
func TestSegmentLengthsAndTruncatedBootstrap(t *testing.T){
 for _,length:=range []int{1500,1666,2000}{
  v:=validVector(length);validateVector(&v)
  var cells [2]Cell;var windows [2][2000][2]Counts
  for age:=range 2{for pos:=range length{for arm:=range 2{cells[age].Blocks[pos][arm]=Counts{C:1};windows[age][pos][arm]=Counts{C:20}}}}
  totals,draws:=commitSample(rng(0,0,"bootstrap/0"),length,&cells,&windows)
  if draws!=(length+19)/20{t.Fatal("wrong circular start count",length,draws)}
  for age:=range 2{for arm:=range 2{if totals[age][arm].C!=length{t.Fatal("not truncated to L",length,totals)}}}
 }
}
func TestSegmentFloorRefused(t *testing.T){v:=validVector(1499);defer func(){if recover()==nil{t.Fatal("1499 accepted")}}();validateVector(&v)}
func TestAgeLengthMismatchRefused(t *testing.T){v:=validVector(1500);v.Counts[0][1]=make([]int,1666);defer func(){if recover()==nil{t.Fatal("age mismatch accepted")}}();validateVector(&v)}
func TestStopReasonsPreserveFirst(t *testing.T){
 status:="pass";reasons:=[]string{};recordStop(&status,&reasons,"not-testable-OC-gate");recordStop(&status,&reasons,"not-testable-projected-compute-budget");recordStop(&status,&reasons,"not-testable-OC-gate")
 if status!="not-testable-OC-gate"||len(reasons)!=2||reasons[1]!="not-testable-projected-compute-budget"{t.Fatal(status,reasons)}
}
func TestConditionalBounds(t *testing.T){
 lo,hi,ok:=combinedBounds(filled(.9,399),filled(.79,399))
 if !ok||lo!=.79||hi!=.9{t.Fatal(lo,hi,ok)}
 _,hi,ok=combinedBounds(filled(.7,399),filled(.81,399))
 if !ok||hi!=.81{t.Fatal("negative must use maximum upper")}
 _,_,ok=combinedBounds(filled(.9,399),filled(.9,378));if ok{t.Fatal("invalid directory accepted")}
 _,_,ok=combinedBounds(nil,nil);if ok{t.Fatal("empty accepted")}
 for _,invalid:=range []float64{math.NaN(),math.Inf(1),1.1,-.1}{_,_,ok=combinedBounds(filled(.9,399),filled(invalid,399));if ok{t.Fatal("invalid bound accepted",invalid)}}
}
func TestEffectiveThresholds(t *testing.T){
 for _,row:=range [][3]int{{256,219,5},{512,429,15},{1024,847,36}}{
  n,p,e:=row[0],row[1],row[2]
  lo,_:=gateInterval(p,n);below,_:=gateInterval(p-1,n)
  _,hi:=gateInterval(e,n);_,above:=gateInterval(e+1,n)
  if lo<.8||below>=.8||hi>.05||above<=.05{t.Fatal(row,lo,below,hi,above)}
 }
}
func TestUnequalDenominators(t *testing.T){
 actual:=ratios(Counts{C:70,A:110,E:90,UA:7,U:20})
 expected:=[4]float64{70./117,70./110,77./117,77./97}
 for i:=range actual{if math.Abs(actual[i]-expected[i])>1e-14{t.Fatal(i,actual,expected)}}
}
func TestVectorMismatchRefused(t *testing.T){
 v:=validVector(1500);v.Counts[0][0][0]=2
 v.DirectoryCases[0][0]=[]DirectoryCase{{Position:0,DirectoryHash:strings.Repeat("a",64)}}
 defer func(){if recover()==nil{t.Fatal("count mismatch accepted")}}()
 validateVector(&v)
}
'''
        with tempfile.TemporaryDirectory() as directory:
            base = Path(directory)
            (base / 'oc.go').write_bytes((ROOT / 'oc_v5.go').read_bytes())
            (base / 'oc_test.go').write_text(source)
            environment = dict(recipe['environment'])
            for key in ('HOME', 'GOCACHE', 'GOTMPDIR'):
                environment[key] = str(base / key.lower())
                Path(environment[key]).mkdir()
            argv = [recipe['go']['path'], 'test', '-pgo=off', '-count=1',
                    str(base / 'oc.go'), str(base / 'oc_test.go')]
            completed = subprocess.run(argv, cwd=recipe['cwd'], env=environment,
                                       capture_output=True, text=True, timeout=180)
            self.assertEqual(completed.returncode, 0, completed.stdout + completed.stderr)


if __name__ == '__main__':
    unittest.main()
