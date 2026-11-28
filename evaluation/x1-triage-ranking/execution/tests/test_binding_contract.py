"""Pre-freeze Go/Python contract evidence from shared synthetic records only."""
import copy
import json
from pathlib import Path
import sys
import unittest

EXEC=Path(__file__).resolve().parents[1]
BINDING=EXEC/'binding'
sys.path.insert(0,str(EXEC));sys.path.insert(0,str(BINDING))
from acceptance import verify_bound_policy, verify_model_proof
from interlock import InterlockFailure


class ModelFreeContractTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.golden=json.loads((Path(__file__).with_name('model_free_contract.json')).read_bytes())

    def test_exact_record_routes(self):
        self.assertEqual(set(self.golden),{'effective','receipts'})
        self.assertEqual(set(self.golden['effective']),{'bound','reference','synthetic-unbound'})
        self.assertEqual(set(self.golden['receipts']),{'bound','reference'})
        for collection,records in self.golden.items():
            for route,record in records.items():
                with self.subTest(collection=collection,route=route):
                    self.assertEqual(record['route'],route)
                    self.assertEqual(record['revision'],'contract-revision')

    def test_real_effective_and_receipt_proofs_pass(self):
        self.assertTrue(verify_bound_policy(self.golden['effective']['bound']))
        for route in ('bound','reference'):
            with self.subTest(route=route):
                self.assertTrue(verify_model_proof(self.golden['effective'][route]))
                self.assertTrue(verify_model_proof(self.golden['receipts'][route],route=route))

    def test_missing_model_proof_key_refuses_policies_and_receipts(self):
        for collection in ('effective','receipts'):
            for route in ('bound','reference'):
                record=self.golden[collection][route]
                for key in record['model_proof']:
                    with self.subTest(collection=collection,route=route,key=key):
                        changed=copy.deepcopy(record)
                        del changed['model_proof'][key]
                        with self.assertRaisesRegex(InterlockFailure,'model-call-impossible proof incomplete'):
                            if collection=='effective':
                                verify_model_proof(changed)
                            else:
                                verify_model_proof(changed,route=route)

    def test_flipped_association_nil_refuses_bound_policy(self):
        # Receipts carry association_result, not the effective API-input flag.
        changed=copy.deepcopy(self.golden['effective']['bound'])
        self.assertIs(changed['association_nil'],True)
        changed['association_nil']=not changed['association_nil']
        with self.assertRaisesRegex(InterlockFailure,'effective API inputs violate scanner-only binding'):
            verify_bound_policy(changed)

    def test_actual_lane_reasons_and_structural_absence(self):
        bound=self.golden['receipts']['bound']
        expected={
            'contextual':('skipped','skipped (not opted in)',1),
            'clawpatch-review':('skipped','skipped (not opted in)',1),
            'association':('skipped','disabled',0),
            'association-fallback':('skipped','disabled',0),
        }
        lanes=bound['lane_report']
        self.assertEqual(len(lanes),len(expected))
        self.assertEqual({row['lane'] for row in lanes},set(expected))
        for index,row in enumerate(lanes):
            with self.subTest(lane=row['lane']):
                self.assertEqual((row['status'],row['reason'],row['units']),expected[row['lane']])
                changed=copy.deepcopy(bound)
                changed['lane_report'][index]['reason']='enabled'
                with self.assertRaisesRegex(InterlockFailure,'actual bound model lane not disabled'):
                    verify_model_proof(changed,route='bound')
        reference=self.golden['receipts']['reference']
        self.assertIsNone(reference['lane_report'])
        self.assertEqual(reference['model_stages'],'structurally_absent')
        self.assertEqual(reference['association_result'],'structurally_absent')
        changed=copy.deepcopy(reference)
        changed['lane_report']=[]
        with self.assertRaisesRegex(InterlockFailure,'reference model stages not structurally absent'):
            verify_model_proof(changed,route='reference')

    def test_control_cannot_claim_bound_model_proof(self):
        control=self.golden['effective']['synthetic-unbound']
        self.assertEqual(control['model_stages'],'enabled_control')
        self.assertIs(control['model_call_impossible'],False)
        self.assertTrue(control['model_proof'])
        for key,value in control['model_proof'].items():
            with self.subTest(key=key):
                self.assertIs(value,False)
        with self.assertRaisesRegex(InterlockFailure,'model-call-impossible proof incomplete'):
            verify_model_proof(control)
        with self.assertRaises(InterlockFailure):
            verify_bound_policy(control)


if __name__=='__main__':
    unittest.main()
