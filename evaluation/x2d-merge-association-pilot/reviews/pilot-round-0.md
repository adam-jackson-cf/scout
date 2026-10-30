# X2d — round 0 record and council critique

## Main-agent record (before any label; 2026-10-30)

**Sign-off conditions (council round 11, `reviews/round-11.md` §5 and Addendum).**
1. All 44 frozen-file digests and the `tune_threshold.py` pin were recomputed and equal the round-11 table.
2. `signoff.json` was written (round 11): SHA-256 `e7059f2a9407e38c4d5e8aff0480913fab36a32a69ccbb9721c9633934fdbc2d`. Its `code_under_test` holds 17 rows, equal row for row to the round-9 record. `x2d.verify_pins()` returns `[]`.
3. Validation ran on these bytes:
   - (a) with `signoff.json` moved aside: **651 checks, 0 failed** (`/tmp/x2d-validate-r11-norecord.log`). The record's digest was `e7059f2a…dc2d` both before and after the move.
   - (b) with the record in place: **650 passed, 1 failed**. The failure is `C without the council's sign-off record the collector refuses before any clone or model call`, which is expected per the Addendum (`/tmp/x2d-validate-r11-signoff.log`).
4. Commit `0b5659440c7368ac18b82f60a4edf3956408d2bf`, pushed.
5. The run worktree `../scout-x2d` was moved to `0b56594` by checkout only. `data/` stayed untracked and untouched; the catalogue digest is unchanged.

**Two commits (PREREGISTRATION §14 item 18; for RESULTS `Deviations`).**
- These stages ran at `730674e` under the rev-9 sign-off: onboarding (16 maps, all rc 0), consensus and catalogue freeze (114 areas, sha256 `1220016c26115ccd3e3a521de76ec45e68893f4ab6edf0f4889e8c8ce2c84199`), and quality with the intruder audit (40/40 both labellers, κ 1.0).
- The first `gates` attempt refused with the P2 shortfall (`reviews/pilot-round-0-p2-gate-memo.md`). Nothing after the refusal ran.
- Everything from `gates-build` onward runs at `0b56594`.

**R10-03 comparison: `gates-build` alone (`/tmp/x2d-gates-build.log`) against `reviews/round-10-response.md` §2 and `reviews/round-11.md` §2.** Keys are `(kind, source_case_id)`. The file is in blinded order, so sets are compared, not file order.

| item | recorded | live build | equal |
|---|---|---|---|
| kind counts | p1 20, p2 12, dup 9, planted_new 3, prose 1, trap 1 | same | yes |
| P1 source cluster ids (set) | the 20 ids of round-10-response §2 | same 20 | yes |
| P2 source cluster ids (set) | the 12 ids of round-10-response §2 (4 P unplaceable, 2 P leftover, 6 D) | same 12; stdout `p2_donors` {P_unplaceable 4, P_leftover 2, D 6} | yes |
| `ablated_area_id` set / `ablated_area_ids` | 10 areas `26ac01437eaf`, `73232e7095cf`, `95b28e00298a`, `9a4a0e49b8c6`, `c207aaf8cb60`, `c4f3d5556193`, `d56ca48225f1`, `ea1d02de2a93`, `ea75266c1719`, `f36cd1216200` | same 10, identical list on every P2 row | yes |
| `gap_positive` | 12 of 12 | 12 of 12 | yes |
| shared ablated areas | `26ac01437eaf` (`2c4db772`, `af2f04c8`: P/P path twin); `73232e7095cf` (`d636de1b`, `84f3f93b`) | same | yes |
| `catalogue-p1` / `-p2` / `-n1` areas | 114 / 104 / 114 | 114 / 104 / 114 | yes |
| every P1 gold area in `catalogue-p1` | true | true | yes |
| `units.jsonl` / `units.contrast.jsonl` lines | 96 / 20 | 96 / 20 | yes |
| `data/catalogue/catalogue.sha256` | `1220016c…4199` | file and recomputed digest both `1220016c…4199` | yes |
| stdout | eligible 26, eligible_d 19, p1_unplaceable 4, ablated_areas 10 | same | yes |

**P1 `target_dir` list:** `.github`, `.github/actions/notify-translations`, `.github/actions/people`, `.github/workflows`, `docs_src/app_testing/app_b_an_py310`, `docs_src/bigger_applications/app_an_py39/internal`, `docs_src/settings/app01`, `docs_src/templates`, `fastapi/security`, `tests` (×2), `tests/test_modules_same_name_body` (×2), `tests/test_tutorial` (×2), `tests/test_tutorial/test_advanced_middleware`, `tests/test_tutorial/test_configure_swagger_ui`, `tests/test_tutorial/test_header_params`, `tests/test_tutorial/test_query_params`, `tests/test_tutorial/test_request_forms_and_files`.

**K2/K3 from the live build (stdout line 2, membership only):**
- K2 = 0.585 (B0 unique owner on 193 of 330 scored clusters), below the 0.90 ceiling, so there is room for Jev; K2 for B2 = 0.806.
- K3 = 0.052 (17 stratum-3 clusters).
- `real_none_proxy` = 39; core gate not enabled.

No mismatch, so the resume continues with `gates-label`.
