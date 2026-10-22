### dataset counts
| repo | raw areas T0 | eligible areas (>=3 files) | exist at T-1 | >=5 outcome changes | included | included-area changes (sum) | PRs fetched | PRs linked | areas O4 unmeasured | files (file-level) |
|---|---|---|---|---|---|---|---|---|---|---|
| fastapi | 6 | 3 | 3 | 2 | 2 | 82 | 568 | 53 | 0 | 6 |
| webpack | 38 | 30 | 28 | 11 | 10 | 154 | 665 | 75 | 4 | 11 |
| nest | 62 | 38 | 37 | 26 | 26 | 526 | 2244 | 158 | 3 | 43 |
| caddy | 28 | 17 | 16 | 14 | 14 | 408 | 419 | 207 | 2 | 39 |
| axum | skipped: no T-1 snapshot (history starts after T0-365d) | | | | | | | | | |
| junit5 | 28 | 15 | 14 | 9 | 8 | 131 | 151 | 24 | 7 | 4 |
| eShopOnWeb | 38 | 25 | 25 | 0 | 0 | 0 | 74 | 0 | 0 | 0 |

### area-level comparisons
| outcome | n | comparison | pooled relative MSE reduction | 95% CI | p (1-sided) | repos improved (>=10 units) | rule |
|---|---|---|---|---|---|---|---|
| o1 | 60 | base vs base+level | -0.44% (MSE 0.5829 -> 0.5854) | [-2.72%, +1.42%] | 0.679 | 3/3 (need 2) | no |
| o1 | 60 | base+level vs base+level+degradation | +0.49% (MSE 0.5854 -> 0.5826) | [-2.89%, +4.33%] | 0.432 | 2/3 (need 2) | no |
| o1 | 60 | base vs base+coupling | -1.24% (MSE 0.5829 -> 0.5901) | [-9.08%, +6.04%] | 0.621 | 2/3 (need 2) | no |
| o1 | 60 | base vs base+ownership | -0.39% (MSE 0.5829 -> 0.5851) | [-1.43%, +0.53%] | 0.798 | 2/3 (need 2) | no |
| o2 | 60 | base vs base+level | +3.88% (MSE 0.2265 -> 0.2177) | [-3.03%, +11.44%] | 0.137 | 2/3 (need 2) | no |
| o2 | 60 | base+level vs base+level+degradation | -0.52% (MSE 0.2177 -> 0.2188) | [-1.68%, +0.30%] | 0.869 | 1/3 (need 2) | no |
| o2 | 60 | base vs base+coupling | +0.20% (MSE 0.2265 -> 0.2260) | [-2.53%, +2.48%] | 0.407 | 2/3 (need 2) | no |
| o2 | 60 | base vs base+ownership | -0.27% (MSE 0.2265 -> 0.2271) | [-5.06%, +4.12%] | 0.508 | 2/3 (need 2) | no |
| o3 | 60 | base vs base+level | -0.92% (MSE 0.0413 -> 0.0417) | [-6.33%, +3.54%] | 0.636 | 3/3 (need 2) | no |
| o3 | 60 | base+level vs base+level+degradation | -2.78% (MSE 0.0417 -> 0.0428) | [-23.61%, +13.55%] | 0.607 | 1/3 (need 2) | no |
| o3 | 60 | base vs base+coupling | +1.49% (MSE 0.0413 -> 0.0407) | [-3.23%, +5.68%] | 0.256 | 2/3 (need 2) | no |
| o3 | 60 | base vs base+ownership | -1.73% (MSE 0.0413 -> 0.0420) | [-6.35%, +2.36%] | 0.785 | 0/3 (need 2) | no |
| o4a | 35 | base vs base+level | -0.27% (MSE 0.1174 -> 0.1177) | [-3.55%, +2.73%] | 0.534 | 0/2 (need 1) | no |
| o4a | 35 | base+level vs base+level+degradation | +3.46% (MSE 0.1177 -> 0.1136) | [-1.66%, +8.89%] | 0.093 | 0/2 (need 1) | no |
| o4a | 35 | base vs base+coupling | +2.39% (MSE 0.1174 -> 0.1146) | [-4.09%, +8.56%] | 0.239 | 2/2 (need 1) | no |
| o4a | 35 | base vs base+ownership | -2.70% (MSE 0.1174 -> 0.1206) | [-7.28%, +1.83%] | 0.868 | 1/2 (need 1) | no |
| o4b | 35 | base vs base+level | +0.21% (MSE 0.3375 -> 0.3368) | [-8.82%, +7.60%] | 0.460 | 1/2 (need 1) | no |
| o4b | 35 | base+level vs base+level+degradation | -4.09% (MSE 0.3368 -> 0.3506) | [-11.24%, +2.28%] | 0.893 | 1/2 (need 1) | no |
| o4b | 35 | base vs base+coupling | -2.18% (MSE 0.3375 -> 0.3449) | [-4.55%, +0.07%] | 0.968 | 0/2 (need 1) | no |
| o4b | 35 | base vs base+ownership | -1.08% (MSE 0.3375 -> 0.3412) | [-2.45%, +0.49%] | 0.909 | 0/2 (need 1) | no |
| o4c | 35 | base vs base+level | +9.05% (MSE 2.4489 -> 2.2272) | [-2.89%, +18.04%] | 0.058 | 1/2 (need 1) | no |
| o4c | 35 | base+level vs base+level+degradation | +1.08% (MSE 2.2272 -> 2.2031) | [-6.41%, +7.30%] | 0.372 | 1/2 (need 1) | no |
| o4c | 35 | base vs base+coupling | +4.00% (MSE 2.4489 -> 2.3508) | [-1.61%, +9.39%] | 0.077 | 2/2 (need 1) | no |
| o4c | 35 | base vs base+ownership | -3.16% (MSE 2.4489 -> 2.5263) | [-6.22%, +0.18%] | 0.963 | 0/2 (need 1) | no |
| o5a | 60 | base vs base+level | +4.08% (MSE 0.5185 -> 0.4973) | [-9.70%, +14.37%] | 0.287 | 2/3 (need 2) | no |
| o5a | 60 | base+level vs base+level+degradation | +2.13% (MSE 0.4973 -> 0.4867) | [-6.55%, +14.41%] | 0.327 | 1/3 (need 2) | no |
| o5a | 60 | base vs base+coupling | +1.10% (MSE 0.5185 -> 0.5128) | [-3.10%, +5.03%] | 0.281 | 3/3 (need 2) | no |
| o5a | 60 | base vs base+ownership | +1.25% (MSE 0.5185 -> 0.5120) | [-2.37%, +5.71%] | 0.244 | 2/3 (need 2) | no |
| o5b | 35 | base vs base+level | -5.43% (MSE 0.8846 -> 0.9327) | [-12.97%, +0.77%] | 0.962 | 1/2 (need 1) | no |
| o5b | 35 | base+level vs base+level+degradation | +2.78% (MSE 0.9327 -> 0.9067) | [-0.38%, +6.32%] | 0.046 | 2/2 (need 1) | no |
| o5b | 35 | base vs base+coupling | +0.06% (MSE 0.8846 -> 0.8841) | [-0.17%, +0.27%] | 0.317 | 0/2 (need 1) | no |
| o5b | 35 | base vs base+ownership | +0.06% (MSE 0.8846 -> 0.8841) | [-0.17%, +0.27%] | 0.326 | 0/2 (need 1) | no |

### area-level ablations (baseline + one predictor)
| outcome | predictor | reduction | 95% CI | p | repos | rule |
|---|---|---|---|---|---|---|
| o1 | erosion | +1.69% (MSE 0.5829 -> 0.5730) | [-3.49%, +6.35%] | 0.264 | 3/3 (need 2) | no |
| o1 | log_n_high_cc | -0.21% (MSE 0.5829 -> 0.5841) | [-1.13%, +0.67%] | 0.684 | 3/3 (need 2) | no |
| o1 | clone_pct | -0.15% (MSE 0.5829 -> 0.5837) | [-0.74%, +0.38%] | 0.678 | 0/3 (need 2) | no |
| o1 | hotspot | -0.60% (MSE 0.5829 -> 0.5863) | [-2.40%, +0.64%] | 0.809 | 3/3 (need 2) | no |
| o1 | d_erosion | +3.00% (MSE 0.5829 -> 0.5654) | [-0.80%, +7.14%] | 0.063 | 3/3 (need 2) | no |
| o1 | d_log_n_high_cc | -0.37% (MSE 0.5829 -> 0.5850) | [-1.89%, +1.42%] | 0.668 | 2/3 (need 2) | no |
| o1 | d_clone_pct | -0.83% (MSE 0.5829 -> 0.5877) | [-2.03%, +0.78%] | 0.863 | 2/3 (need 2) | no |
| o1 | coupling | -1.24% (MSE 0.5829 -> 0.5901) | [-9.08%, +6.04%] | 0.621 | 2/3 (need 2) | no |
| o1 | ownership | -0.39% (MSE 0.5829 -> 0.5851) | [-1.43%, +0.53%] | 0.798 | 2/3 (need 2) | no |
| o2 | erosion | +1.39% (MSE 0.2265 -> 0.2233) | [-0.86%, +3.73%] | 0.119 | 1/3 (need 2) | no |
| o2 | log_n_high_cc | +1.05% (MSE 0.2265 -> 0.2241) | [-1.99%, +3.89%] | 0.248 | 2/3 (need 2) | no |
| o2 | clone_pct | +2.86% (MSE 0.2265 -> 0.2200) | [-3.89%, +10.23%] | 0.212 | 1/3 (need 2) | no |
| o2 | hotspot | -0.60% (MSE 0.2265 -> 0.2278) | [-2.08%, +0.67%] | 0.835 | 1/3 (need 2) | no |
| o2 | d_erosion | -0.44% (MSE 0.2265 -> 0.2275) | [-1.32%, +0.22%] | 0.877 | 0/3 (need 2) | no |
| o2 | d_log_n_high_cc | -0.60% (MSE 0.2265 -> 0.2278) | [-2.70%, +1.84%] | 0.730 | 1/3 (need 2) | no |
| o2 | d_clone_pct | +0.00% (MSE 0.2265 -> 0.2265) | [-0.06%, +0.06%] | 0.503 | 1/3 (need 2) | no |
| o2 | coupling | +0.20% (MSE 0.2265 -> 0.2260) | [-2.53%, +2.48%] | 0.407 | 2/3 (need 2) | no |
| o2 | ownership | -0.27% (MSE 0.2265 -> 0.2271) | [-5.06%, +4.12%] | 0.508 | 2/3 (need 2) | no |
| o3 | erosion | -1.75% (MSE 0.0413 -> 0.0420) | [-7.09%, +3.02%] | 0.768 | 2/3 (need 2) | no |
| o3 | log_n_high_cc | +1.02% (MSE 0.0413 -> 0.0409) | [-3.71%, +4.73%] | 0.303 | 3/3 (need 2) | no |
| o3 | clone_pct | -0.99% (MSE 0.0413 -> 0.0417) | [-4.00%, +2.33%] | 0.724 | 1/3 (need 2) | no |
| o3 | hotspot | -1.07% (MSE 0.0413 -> 0.0417) | [-4.40%, +1.84%] | 0.747 | 2/3 (need 2) | no |
| o3 | d_erosion | +0.65% (MSE 0.0413 -> 0.0410) | [-6.33%, +7.02%] | 0.397 | 2/3 (need 2) | no |
| o3 | d_log_n_high_cc | +0.04% (MSE 0.0413 -> 0.0413) | [-4.39%, +4.10%] | 0.478 | 2/3 (need 2) | no |
| o3 | d_clone_pct | -0.69% (MSE 0.0413 -> 0.0416) | [-4.37%, +2.37%] | 0.629 | 2/3 (need 2) | no |
| o3 | coupling | +1.49% (MSE 0.0413 -> 0.0407) | [-3.23%, +5.68%] | 0.256 | 2/3 (need 2) | no |
| o3 | ownership | -1.73% (MSE 0.0413 -> 0.0420) | [-6.35%, +2.36%] | 0.785 | 0/3 (need 2) | no |
| o4a | erosion | -0.15% (MSE 0.1174 -> 0.1176) | [-0.51%, +0.18%] | 0.828 | 1/2 (need 1) | no |
| o4a | log_n_high_cc | +0.41% (MSE 0.1174 -> 0.1169) | [-0.25%, +1.29%] | 0.155 | 0/2 (need 1) | no |
| o4a | clone_pct | +0.61% (MSE 0.1174 -> 0.1167) | [-1.38%, +2.54%] | 0.270 | 1/2 (need 1) | no |
| o4a | hotspot | -0.21% (MSE 0.1174 -> 0.1176) | [-0.41%, -0.04%] | 0.995 | 0/2 (need 1) | no |
| o4a | d_erosion | +0.36% (MSE 0.1174 -> 0.1170) | [-0.22%, +1.00%] | 0.117 | 1/2 (need 1) | no |
| o4a | d_log_n_high_cc | +0.00% (MSE 0.1174 -> 0.1174) | [-0.00%, +0.01%] | 0.087 | 1/2 (need 1) | no |
| o4a | d_clone_pct | +1.82% (MSE 0.1174 -> 0.1153) | [-1.15%, +5.18%] | 0.130 | 0/2 (need 1) | no |
| o4a | coupling | +2.39% (MSE 0.1174 -> 0.1146) | [-4.09%, +8.56%] | 0.239 | 2/2 (need 1) | no |
| o4a | ownership | -2.70% (MSE 0.1174 -> 0.1206) | [-7.28%, +1.83%] | 0.868 | 1/2 (need 1) | no |
| o4b | erosion | -0.43% (MSE 0.3375 -> 0.3390) | [-12.56%, +8.99%] | 0.521 | 1/2 (need 1) | no |
| o4b | log_n_high_cc | +0.43% (MSE 0.3375 -> 0.3361) | [-2.93%, +3.50%] | 0.381 | 2/2 (need 1) | no |
| o4b | clone_pct | -0.36% (MSE 0.3375 -> 0.3387) | [-2.96%, +2.21%] | 0.585 | 1/2 (need 1) | no |
| o4b | hotspot | +0.13% (MSE 0.3375 -> 0.3371) | [-3.26%, +3.39%] | 0.430 | 2/2 (need 1) | no |
| o4b | d_erosion | -1.04% (MSE 0.3375 -> 0.3410) | [-5.67%, +2.23%] | 0.668 | 1/2 (need 1) | no |
| o4b | d_log_n_high_cc | -2.80% (MSE 0.3375 -> 0.3469) | [-16.04%, +4.57%] | 0.654 | 1/2 (need 1) | no |
| o4b | d_clone_pct | +3.13% (MSE 0.3375 -> 0.3269) | [-0.45%, +6.16%] | 0.043 | 1/2 (need 1) | no |
| o4b | coupling | -2.18% (MSE 0.3375 -> 0.3449) | [-4.55%, +0.07%] | 0.968 | 0/2 (need 1) | no |
| o4b | ownership | -1.08% (MSE 0.3375 -> 0.3412) | [-2.45%, +0.49%] | 0.909 | 0/2 (need 1) | no |
| o4c | erosion | +6.55% (MSE 2.4489 -> 2.2885) | [-1.96%, +13.46%] | 0.068 | 1/2 (need 1) | no |
| o4c | log_n_high_cc | +1.21% (MSE 2.4489 -> 2.4193) | [-4.48%, +6.93%] | 0.282 | 2/2 (need 1) | no |
| o4c | clone_pct | -0.41% (MSE 2.4489 -> 2.4589) | [-1.52%, +0.85%] | 0.749 | 1/2 (need 1) | no |
| o4c | hotspot | +7.34% (MSE 2.4489 -> 2.2693) | [-3.27%, +16.06%] | 0.072 | 1/2 (need 1) | no |
| o4c | d_erosion | +8.54% (MSE 2.4489 -> 2.2399) | [-5.77%, +20.38%] | 0.104 | 2/2 (need 1) | no |
| o4c | d_log_n_high_cc | +7.48% (MSE 2.4489 -> 2.2657) | [-3.80%, +16.12%] | 0.093 | 2/2 (need 1) | no |
| o4c | d_clone_pct | +3.96% (MSE 2.4489 -> 2.3518) | [-1.43%, +7.93%] | 0.067 | 1/2 (need 1) | no |
| o4c | coupling | +4.00% (MSE 2.4489 -> 2.3508) | [-1.61%, +9.39%] | 0.077 | 2/2 (need 1) | no |
| o4c | ownership | -3.16% (MSE 2.4489 -> 2.5263) | [-6.22%, +0.18%] | 0.963 | 0/2 (need 1) | no |
| o5a | erosion | +2.14% (MSE 0.5185 -> 0.5074) | [-1.60%, +6.70%] | 0.128 | 3/3 (need 2) | no |
| o5a | log_n_high_cc | +1.80% (MSE 0.5185 -> 0.5091) | [-1.98%, +5.71%] | 0.156 | 3/3 (need 2) | no |
| o5a | clone_pct | +1.41% (MSE 0.5185 -> 0.5112) | [-13.70%, +13.10%] | 0.436 | 1/3 (need 2) | no |
| o5a | hotspot | +0.11% (MSE 0.5185 -> 0.5179) | [-1.62%, +1.45%] | 0.437 | 2/3 (need 2) | no |
| o5a | d_erosion | +0.01% (MSE 0.5185 -> 0.5185) | [-0.98%, +0.97%] | 0.499 | 2/3 (need 2) | no |
| o5a | d_log_n_high_cc | -1.72% (MSE 0.5185 -> 0.5274) | [-4.02%, -0.40%] | 0.996 | 1/3 (need 2) | no |
| o5a | d_clone_pct | -4.24% (MSE 0.5185 -> 0.5404) | [-10.90%, +0.92%] | 0.953 | 1/3 (need 2) | no |
| o5a | coupling | +1.10% (MSE 0.5185 -> 0.5128) | [-3.10%, +5.03%] | 0.281 | 3/3 (need 2) | no |
| o5a | ownership | +1.25% (MSE 0.5185 -> 0.5120) | [-2.37%, +5.71%] | 0.244 | 2/3 (need 2) | no |
| o5b | erosion | -2.50% (MSE 0.8846 -> 0.9068) | [-6.04%, +1.20%] | 0.915 | 1/2 (need 1) | no |
| o5b | log_n_high_cc | -2.40% (MSE 0.8846 -> 0.9059) | [-5.31%, +0.23%] | 0.961 | 0/2 (need 1) | no |
| o5b | clone_pct | +0.06% (MSE 0.8846 -> 0.8841) | [-0.17%, +0.27%] | 0.321 | 0/2 (need 1) | no |
| o5b | hotspot | -3.51% (MSE 0.8846 -> 0.9157) | [-7.43%, -0.13%] | 0.977 | 1/2 (need 1) | no |
| o5b | d_erosion | +0.06% (MSE 0.8846 -> 0.8841) | [-1.49%, +2.35%] | 0.492 | 2/2 (need 1) | no |
| o5b | d_log_n_high_cc | -5.44% (MSE 0.8846 -> 0.9327) | [-17.04%, +2.46%] | 0.905 | 1/2 (need 1) | no |
| o5b | d_clone_pct | +0.06% (MSE 0.8846 -> 0.8841) | [-0.17%, +0.27%] | 0.314 | 0/2 (need 1) | no |
| o5b | coupling | +0.06% (MSE 0.8846 -> 0.8841) | [-0.17%, +0.27%] | 0.317 | 0/2 (need 1) | no |
| o5b | ownership | +0.06% (MSE 0.8846 -> 0.8841) | [-0.17%, +0.27%] | 0.326 | 0/2 (need 1) | no |

### file-level comparisons
| outcome | n | comparison | pooled relative MSE reduction | 95% CI | p (1-sided) | repos improved (>=10 units) | rule |
|---|---|---|---|---|---|---|---|
| o1 | 103 | base vs base+level | -5.41% (MSE 0.4152 -> 0.4377) | [-12.34%, +1.38%] | 0.945 | 1/3 (need 2) | no |
| o1 | 103 | base+level vs base+level+degradation | -2.54% (MSE 0.4377 -> 0.4488) | [-7.37%, +2.49%] | 0.838 | 0/3 (need 2) | no |
| o1 | 103 | base vs base+coupling | -0.48% (MSE 0.4152 -> 0.4172) | [-3.79%, +2.59%] | 0.615 | 1/3 (need 2) | no |
| o1 | 103 | base vs base+ownership | +2.94% (MSE 0.4152 -> 0.4030) | [-2.61%, +8.02%] | 0.132 | 2/3 (need 2) | no |
| o2 | 103 | base vs base+level | -0.52% (MSE 0.2081 -> 0.2092) | [-2.87%, +1.79%] | 0.680 | 1/3 (need 2) | no |
| o2 | 103 | base+level vs base+level+degradation | -1.31% (MSE 0.2092 -> 0.2120) | [-5.14%, +2.20%] | 0.772 | 0/3 (need 2) | no |
| o2 | 103 | base vs base+coupling | -0.34% (MSE 0.2081 -> 0.2088) | [-3.77%, +2.60%] | 0.582 | 2/3 (need 2) | no |
| o2 | 103 | base vs base+ownership | +3.64% (MSE 0.2081 -> 0.2006) | [+0.77%, +6.31%] | 0.011 | 3/3 (need 2) | no |
| o3 | 103 | base vs base+level | -1.91% (MSE 0.0609 -> 0.0620) | [-7.46%, +1.33%] | 0.751 | 1/3 (need 2) | no |
| o3 | 103 | base+level vs base+level+degradation | +4.01% (MSE 0.0620 -> 0.0596) | [-0.51%, +10.70%] | 0.055 | 3/3 (need 2) | no |
| o3 | 103 | base vs base+coupling | +0.63% (MSE 0.0609 -> 0.0605) | [-2.30%, +3.60%] | 0.349 | 3/3 (need 2) | no |
| o3 | 103 | base vs base+ownership | -1.72% (MSE 0.0609 -> 0.0619) | [-6.43%, +2.91%] | 0.764 | 1/3 (need 2) | no |
| o5a | 103 | base vs base+level | +5.91% (MSE 0.4332 -> 0.4076) | [+1.66%, +11.23%] | 0.004 | 1/3 (need 2) | no |
| o5a | 103 | base+level vs base+level+degradation | -2.69% (MSE 0.4076 -> 0.4186) | [-6.14%, -0.65%] | 1.000 | 0/3 (need 2) | no |
| o5a | 103 | base vs base+coupling | +0.08% (MSE 0.4332 -> 0.4329) | [-0.88%, +0.94%] | 0.405 | 3/3 (need 2) | no |
| o5a | 103 | base vs base+ownership | +0.45% (MSE 0.4332 -> 0.4313) | [-1.07%, +1.92%] | 0.284 | 3/3 (need 2) | no |

### file-level ablations (baseline + one predictor)
| outcome | predictor | reduction | 95% CI | p | repos | rule |
|---|---|---|---|---|---|---|
| o1 | erosion | -0.03% (MSE 0.4152 -> 0.4153) | [-1.71%, +1.29%] | 0.497 | 3/3 (need 2) | no |
| o1 | log_n_high_cc | -0.22% (MSE 0.4152 -> 0.4161) | [-2.72%, +1.63%] | 0.583 | 3/3 (need 2) | no |
| o1 | clone_pct | -4.17% (MSE 0.4152 -> 0.4325) | [-7.56%, -1.56%] | 0.999 | 2/3 (need 2) | no |
| o1 | hotspot | -0.06% (MSE 0.4152 -> 0.4155) | [-2.78%, +2.72%] | 0.538 | 2/3 (need 2) | no |
| o1 | d_erosion | -2.71% (MSE 0.4152 -> 0.4265) | [-4.56%, -1.29%] | 1.000 | 0/3 (need 2) | no |
| o1 | d_log_n_high_cc | -1.62% (MSE 0.4152 -> 0.4219) | [-3.31%, -0.28%] | 0.988 | 0/3 (need 2) | no |
| o1 | d_clone_pct | -2.78% (MSE 0.4152 -> 0.4268) | [-10.05%, +4.38%] | 0.779 | 1/3 (need 2) | no |
| o1 | coupling | -0.48% (MSE 0.4152 -> 0.4172) | [-3.79%, +2.59%] | 0.615 | 1/3 (need 2) | no |
| o1 | ownership | +2.94% (MSE 0.4152 -> 0.4030) | [-2.61%, +8.02%] | 0.132 | 2/3 (need 2) | no |
| o2 | erosion | -0.35% (MSE 0.2081 -> 0.2089) | [-1.23%, +0.54%] | 0.778 | 1/3 (need 2) | no |
| o2 | log_n_high_cc | +0.05% (MSE 0.2081 -> 0.2080) | [-0.42%, +0.54%] | 0.441 | 2/3 (need 2) | no |
| o2 | clone_pct | -1.21% (MSE 0.2081 -> 0.2106) | [-2.24%, -0.19%] | 0.986 | 0/3 (need 2) | no |
| o2 | hotspot | -0.81% (MSE 0.2081 -> 0.2098) | [-2.04%, +0.37%] | 0.918 | 1/3 (need 2) | no |
| o2 | d_erosion | -0.88% (MSE 0.2081 -> 0.2099) | [-2.52%, +0.75%] | 0.853 | 1/3 (need 2) | no |
| o2 | d_log_n_high_cc | -3.39% (MSE 0.2081 -> 0.2152) | [-7.72%, +0.69%] | 0.944 | 0/3 (need 2) | no |
| o2 | d_clone_pct | -1.48% (MSE 0.2081 -> 0.2112) | [-4.35%, +1.38%] | 0.843 | 0/3 (need 2) | no |
| o2 | coupling | -0.34% (MSE 0.2081 -> 0.2088) | [-3.77%, +2.60%] | 0.582 | 2/3 (need 2) | no |
| o2 | ownership | +3.64% (MSE 0.2081 -> 0.2006) | [+0.77%, +6.31%] | 0.011 | 3/3 (need 2) | no |
| o3 | erosion | -1.80% (MSE 0.0609 -> 0.0620) | [-6.55%, +0.90%] | 0.771 | 0/3 (need 2) | no |
| o3 | log_n_high_cc | -0.85% (MSE 0.0609 -> 0.0614) | [-3.04%, +0.66%] | 0.811 | 2/3 (need 2) | no |
| o3 | clone_pct | -0.98% (MSE 0.0609 -> 0.0615) | [-3.63%, +0.98%] | 0.793 | 1/3 (need 2) | no |
| o3 | hotspot | -0.98% (MSE 0.0609 -> 0.0615) | [-4.43%, +1.75%] | 0.699 | 2/3 (need 2) | no |
| o3 | d_erosion | +0.58% (MSE 0.0609 -> 0.0605) | [-0.88%, +2.73%] | 0.300 | 2/3 (need 2) | no |
| o3 | d_log_n_high_cc | -0.16% (MSE 0.0609 -> 0.0610) | [-0.75%, +0.18%] | 0.670 | 0/3 (need 2) | no |
| o3 | d_clone_pct | +0.64% (MSE 0.0609 -> 0.0605) | [-2.32%, +4.48%] | 0.387 | 3/3 (need 2) | no |
| o3 | coupling | +0.63% (MSE 0.0609 -> 0.0605) | [-2.30%, +3.60%] | 0.349 | 3/3 (need 2) | no |
| o3 | ownership | -1.72% (MSE 0.0609 -> 0.0619) | [-6.43%, +2.91%] | 0.764 | 1/3 (need 2) | no |
| o5a | erosion | +3.06% (MSE 0.4332 -> 0.4199) | [+0.81%, +5.98%] | 0.005 | 1/3 (need 2) | no |
| o5a | log_n_high_cc | +3.38% (MSE 0.4332 -> 0.4186) | [+1.12%, +6.26%] | 0.001 | 3/3 (need 2) | no |
| o5a | clone_pct | +6.45% (MSE 0.4332 -> 0.4053) | [+2.02%, +11.85%] | 0.003 | 2/3 (need 2) | **yes** |
| o5a | hotspot | +0.31% (MSE 0.4332 -> 0.4319) | [+0.10%, +0.60%] | 0.002 | 1/3 (need 2) | no |
| o5a | d_erosion | -0.57% (MSE 0.4332 -> 0.4357) | [-1.74%, +0.37%] | 0.869 | 0/3 (need 2) | no |
| o5a | d_log_n_high_cc | -0.01% (MSE 0.4332 -> 0.4333) | [-0.14%, +0.12%] | 0.551 | 0/3 (need 2) | no |
| o5a | d_clone_pct | -5.67% (MSE 0.4332 -> 0.4578) | [-14.61%, -0.91%] | 1.000 | 0/3 (need 2) | no |
| o5a | coupling | +0.08% (MSE 0.4332 -> 0.4329) | [-0.88%, +0.94%] | 0.405 | 3/3 (need 2) | no |
| o5a | ownership | +0.45% (MSE 0.4332 -> 0.4313) | [-1.07%, +1.92%] | 0.284 | 3/3 (need 2) | no |

### per-repo relative MSE reduction (area level)
| outcome | comparison | fastapi | webpack | nest | caddy | axum | junit5 | eShopOnWeb |
|---|---|---|---|---|---|---|---|---|
| o1 (n by repo {'caddy': 14, 'fastapi': 2, 'junit5': 8, 'nest': 26, 'webpack': 10}) | level | +4.84% | +2.18% | +0.00% | +0.01% | - | -7.41% | - |
| o1 (n by repo {'caddy': 14, 'fastapi': 2, 'junit5': 8, 'nest': 26, 'webpack': 10}) | degradation | -7.04% | +0.98% | +0.00% | -0.00% | - | +2.76% | - |
| o1 (n by repo {'caddy': 14, 'fastapi': 2, 'junit5': 8, 'nest': 26, 'webpack': 10}) | coupling | -9.57% | +0.31% | -3.43% | +0.00% | - | +1.22% | - |
| o1 (n by repo {'caddy': 14, 'fastapi': 2, 'junit5': 8, 'nest': 26, 'webpack': 10}) | ownership | -4.93% | -0.61% | +0.00% | +0.00% | - | -0.04% | - |
| o2 (n by repo {'caddy': 14, 'fastapi': 2, 'junit5': 8, 'nest': 26, 'webpack': 10}) | level | +0.07% | +22.48% | -0.00% | +2.12% | - | -0.00% | - |
| o2 (n by repo {'caddy': 14, 'fastapi': 2, 'junit5': 8, 'nest': 26, 'webpack': 10}) | degradation | -0.00% | -3.96% | -0.00% | +0.00% | - | -0.00% | - |
| o2 (n by repo {'caddy': 14, 'fastapi': 2, 'junit5': 8, 'nest': 26, 'webpack': 10}) | coupling | -0.93% | +0.55% | +0.51% | -2.20% | - | +0.00% | - |
| o2 (n by repo {'caddy': 14, 'fastapi': 2, 'junit5': 8, 'nest': 26, 'webpack': 10}) | ownership | +0.07% | -6.43% | +0.30% | +4.91% | - | +2.12% | - |
| o3 (n by repo {'caddy': 14, 'fastapi': 2, 'junit5': 8, 'nest': 26, 'webpack': 10}) | level | -25.61% | +1.05% | +0.01% | +3.79% | - | -5.84% | - |
| o3 (n by repo {'caddy': 14, 'fastapi': 2, 'junit5': 8, 'nest': 26, 'webpack': 10}) | degradation | +25.63% | -7.66% | +0.01% | -14.11% | - | +3.64% | - |
| o3 (n by repo {'caddy': 14, 'fastapi': 2, 'junit5': 8, 'nest': 26, 'webpack': 10}) | coupling | +24.89% | -0.79% | +0.00% | +6.15% | - | +1.63% | - |
| o3 (n by repo {'caddy': 14, 'fastapi': 2, 'junit5': 8, 'nest': 26, 'webpack': 10}) | ownership | -37.96% | -1.99% | -0.00% | -0.79% | - | -1.50% | - |
| o4a (n by repo {'caddy': 11, 'fastapi': 2, 'junit5': 1, 'nest': 15, 'webpack': 6}) | level | +0.00% | -0.83% | -0.01% | -0.00% | - | +0.00% | - |
| o4a (n by repo {'caddy': 11, 'fastapi': 2, 'junit5': 1, 'nest': 15, 'webpack': 6}) | degradation | +0.00% | +10.59% | -0.01% | -0.00% | - | +0.00% | - |
| o4a (n by repo {'caddy': 11, 'fastapi': 2, 'junit5': 1, 'nest': 15, 'webpack': 6}) | coupling | -19.66% | +11.09% | +0.01% | +0.01% | - | +0.00% | - |
| o4a (n by repo {'caddy': 11, 'fastapi': 2, 'junit5': 1, 'nest': 15, 'webpack': 6}) | ownership | +0.00% | -8.31% | +0.00% | -0.00% | - | +0.00% | - |
| o4b (n by repo {'caddy': 11, 'fastapi': 2, 'junit5': 1, 'nest': 15, 'webpack': 6}) | level | +2.39% | +0.01% | +0.01% | -0.44% | - | +0.00% | - |
| o4b (n by repo {'caddy': 11, 'fastapi': 2, 'junit5': 1, 'nest': 15, 'webpack': 6}) | degradation | -2.44% | +0.00% | +0.00% | -16.60% | - | +0.00% | - |
| o4b (n by repo {'caddy': 11, 'fastapi': 2, 'junit5': 1, 'nest': 15, 'webpack': 6}) | coupling | +0.00% | -0.00% | -0.00% | -9.59% | - | +0.00% | - |
| o4b (n by repo {'caddy': 11, 'fastapi': 2, 'junit5': 1, 'nest': 15, 'webpack': 6}) | ownership | -0.00% | +0.00% | -0.00% | -4.77% | - | +0.00% | - |
| o4c (n by repo {'caddy': 11, 'fastapi': 2, 'junit5': 1, 'nest': 15, 'webpack': 6}) | level | -0.73% | +0.02% | -4.48% | +24.10% | - | +0.00% | - |
| o4c (n by repo {'caddy': 11, 'fastapi': 2, 'junit5': 1, 'nest': 15, 'webpack': 6}) | degradation | -0.22% | +0.01% | -3.29% | +4.86% | - | +0.00% | - |
| o4c (n by repo {'caddy': 11, 'fastapi': 2, 'junit5': 1, 'nest': 15, 'webpack': 6}) | coupling | +0.05% | -0.00% | +7.24% | +7.34% | - | +0.00% | - |
| o4c (n by repo {'caddy': 11, 'fastapi': 2, 'junit5': 1, 'nest': 15, 'webpack': 6}) | ownership | +0.05% | -0.00% | -6.77% | -5.50% | - | +0.00% | - |
| o5a (n by repo {'caddy': 14, 'fastapi': 2, 'junit5': 8, 'nest': 26, 'webpack': 10}) | level | -72.71% | +9.60% | -1.06% | +18.69% | - | +5.24% | - |
| o5a (n by repo {'caddy': 14, 'fastapi': 2, 'junit5': 8, 'nest': 26, 'webpack': 10}) | degradation | +28.89% | +23.97% | -2.29% | -4.60% | - | -0.11% | - |
| o5a (n by repo {'caddy': 14, 'fastapi': 2, 'junit5': 8, 'nest': 26, 'webpack': 10}) | coupling | -49.20% | +3.32% | +0.11% | +3.89% | - | +0.59% | - |
| o5a (n by repo {'caddy': 14, 'fastapi': 2, 'junit5': 8, 'nest': 26, 'webpack': 10}) | ownership | +82.85% | +6.04% | +1.13% | -5.23% | - | +0.37% | - |
| o5b (n by repo {'caddy': 11, 'fastapi': 2, 'junit5': 1, 'nest': 15, 'webpack': 6}) | level | +0.00% | +0.20% | +0.00% | -13.35% | - | +0.00% | - |
| o5b (n by repo {'caddy': 11, 'fastapi': 2, 'junit5': 1, 'nest': 15, 'webpack': 6}) | degradation | -0.00% | +0.01% | +0.01% | +6.28% | - | +0.00% | - |
| o5b (n by repo {'caddy': 11, 'fastapi': 2, 'junit5': 1, 'nest': 15, 'webpack': 6}) | coupling | -0.00% | +0.21% | -0.00% | -0.00% | - | +0.00% | - |
| o5b (n by repo {'caddy': 11, 'fastapi': 2, 'junit5': 1, 'nest': 15, 'webpack': 6}) | ownership | +0.00% | +0.21% | -0.01% | -0.00% | - | +0.00% | - |

### bootstrap coefficients (area level, z-scored ridge)
| outcome | model | predictor | coef | 95% CI | construct-invalid |
|---|---|---|---|---|---|
| o1 | coefs_level_model | erosion | -0.045 | [-0.098, -0.003] | False |
| o1 | coefs_level_model | log_n_high_cc | +0.016 | [-0.021, +0.058] | False |
| o1 | coefs_level_model | clone_pct | -0.013 | [-0.069, +0.051] | False |
| o1 | coefs_level_model | hotspot | +0.009 | [-0.031, +0.057] | False |
| o1 | coefs_full_model | erosion | -0.019 | [-0.043, +0.000] | False |
| o1 | coefs_full_model | log_n_high_cc | +0.010 | [-0.007, +0.027] | False |
| o1 | coefs_full_model | clone_pct | -0.005 | [-0.031, +0.022] | False |
| o1 | coefs_full_model | hotspot | +0.006 | [-0.013, +0.027] | False |
| o1 | coefs_full_model | d_erosion | -0.018 | [-0.039, +0.005] | False |
| o1 | coefs_full_model | d_log_n_high_cc | +0.008 | [-0.009, +0.026] | False |
| o1 | coefs_full_model | d_clone_pct | +0.004 | [-0.016, +0.028] | False |
| o1 | coefs_coupling_model | coupling | -0.040 | [-0.097, +0.020] | False |
| o2 | coefs_level_model | erosion | +0.000 | [-0.000, +0.000] | False |
| o2 | coefs_level_model | log_n_high_cc | -0.000 | [-0.000, +0.000] | False |
| o2 | coefs_level_model | clone_pct | +0.000 | [-0.000, +0.000] | False |
| o2 | coefs_level_model | hotspot | -0.000 | [-0.000, +0.000] | False |
| o2 | coefs_full_model | erosion | +0.000 | [-0.000, +0.000] | False |
| o2 | coefs_full_model | log_n_high_cc | -0.000 | [-0.000, +0.000] | False |
| o2 | coefs_full_model | clone_pct | +0.000 | [-0.000, +0.000] | False |
| o2 | coefs_full_model | hotspot | -0.000 | [-0.000, +0.000] | False |
| o2 | coefs_full_model | d_erosion | +0.000 | [-0.000, +0.000] | False |
| o2 | coefs_full_model | d_log_n_high_cc | -0.000 | [-0.000, +0.000] | False |
| o2 | coefs_full_model | d_clone_pct | +0.000 | [-0.000, +0.000] | False |
| o2 | coefs_coupling_model | coupling | +0.013 | [-0.005, +0.028] | False |
| o3 | coefs_level_model | erosion | +0.010 | [-0.026, +0.058] | False |
| o3 | coefs_level_model | log_n_high_cc | +0.037 | [+0.002, +0.073] | False |
| o3 | coefs_level_model | clone_pct | -0.009 | [-0.041, +0.022] | False |
| o3 | coefs_level_model | hotspot | +0.008 | [-0.031, +0.046] | False |
| o3 | coefs_full_model | erosion | +0.021 | [-0.015, +0.064] | False |
| o3 | coefs_full_model | log_n_high_cc | +0.035 | [-0.003, +0.072] | False |
| o3 | coefs_full_model | clone_pct | -0.012 | [-0.044, +0.018] | False |
| o3 | coefs_full_model | hotspot | +0.012 | [-0.025, +0.048] | False |
| o3 | coefs_full_model | d_erosion | -0.037 | [-0.077, +0.005] | False |
| o3 | coefs_full_model | d_log_n_high_cc | -0.003 | [-0.046, +0.044] | False |
| o3 | coefs_full_model | d_clone_pct | -0.008 | [-0.038, +0.031] | False |
| o3 | coefs_coupling_model | coupling | -0.022 | [-0.065, +0.021] | False |
| o4a | coefs_level_model | erosion | +0.000 | [-0.000, +0.000] | False |
| o4a | coefs_level_model | log_n_high_cc | +0.000 | [-0.000, +0.000] | False |
| o4a | coefs_level_model | clone_pct | +0.000 | [-0.000, +0.000] | False |
| o4a | coefs_level_model | hotspot | +0.000 | [-0.000, +0.000] | False |
| o4a | coefs_full_model | erosion | +0.000 | [-0.000, +0.000] | False |
| o4a | coefs_full_model | log_n_high_cc | +0.000 | [-0.000, +0.000] | False |
| o4a | coefs_full_model | clone_pct | +0.000 | [-0.000, +0.000] | False |
| o4a | coefs_full_model | hotspot | +0.000 | [-0.000, +0.000] | False |
| o4a | coefs_full_model | d_erosion | -0.000 | [-0.000, +0.000] | False |
| o4a | coefs_full_model | d_log_n_high_cc | +0.000 | [-0.000, +0.000] | False |
| o4a | coefs_full_model | d_clone_pct | +0.000 | [-0.000, +0.000] | False |
| o4a | coefs_coupling_model | coupling | +0.026 | [-0.004, +0.052] | False |
| o4b | coefs_level_model | erosion | -0.060 | [-0.109, -0.002] | False |
| o4b | coefs_level_model | log_n_high_cc | +0.039 | [-0.008, +0.082] | False |
| o4b | coefs_level_model | clone_pct | +0.031 | [-0.035, +0.096] | False |
| o4b | coefs_level_model | hotspot | +0.026 | [-0.051, +0.102] | False |
| o4b | coefs_full_model | erosion | -0.030 | [-0.052, -0.003] | False |
| o4b | coefs_full_model | log_n_high_cc | +0.018 | [-0.003, +0.036] | False |
| o4b | coefs_full_model | clone_pct | +0.016 | [-0.013, +0.042] | False |
| o4b | coefs_full_model | hotspot | +0.019 | [-0.018, +0.051] | False |
| o4b | coefs_full_model | d_erosion | -0.013 | [-0.050, +0.022] | False |
| o4b | coefs_full_model | d_log_n_high_cc | +0.025 | [-0.004, +0.051] | False |
| o4b | coefs_full_model | d_clone_pct | +0.004 | [-0.034, +0.040] | False |
| o4b | coefs_coupling_model | coupling | +0.000 | [-0.000, +0.000] | False |
| o4c | coefs_level_model | erosion | -0.072 | [-0.169, +0.029] | False |
| o4c | coefs_level_model | log_n_high_cc | +0.060 | [-0.023, +0.138] | False |
| o4c | coefs_level_model | clone_pct | +0.006 | [-0.088, +0.098] | False |
| o4c | coefs_level_model | hotspot | +0.054 | [-0.049, +0.152] | False |
| o4c | coefs_full_model | erosion | -0.064 | [-0.160, +0.038] | False |
| o4c | coefs_full_model | log_n_high_cc | +0.054 | [-0.027, +0.128] | False |
| o4c | coefs_full_model | clone_pct | +0.002 | [-0.080, +0.084] | False |
| o4c | coefs_full_model | hotspot | +0.048 | [-0.052, +0.145] | False |
| o4c | coefs_full_model | d_erosion | -0.052 | [-0.151, +0.044] | False |
| o4c | coefs_full_model | d_log_n_high_cc | +0.070 | [+0.007, +0.139] | False |
| o4c | coefs_full_model | d_clone_pct | -0.038 | [-0.146, +0.071] | False |
| o4c | coefs_coupling_model | coupling | +0.000 | [-0.000, +0.000] | False |
| o5a | coefs_level_model | erosion | +0.045 | [-0.013, +0.112] | False |
| o5a | coefs_level_model | log_n_high_cc | -0.022 | [-0.079, +0.037] | False |
| o5a | coefs_level_model | clone_pct | -0.075 | [-0.153, +0.014] | False |
| o5a | coefs_level_model | hotspot | +0.005 | [-0.065, +0.067] | False |
| o5a | coefs_full_model | erosion | +0.053 | [-0.009, +0.115] | False |
| o5a | coefs_full_model | log_n_high_cc | -0.032 | [-0.089, +0.028] | False |
| o5a | coefs_full_model | clone_pct | -0.079 | [-0.156, +0.012] | False |
| o5a | coefs_full_model | hotspot | +0.001 | [-0.071, +0.062] | False |
| o5a | coefs_full_model | d_erosion | -0.019 | [-0.115, +0.059] | False |
| o5a | coefs_full_model | d_log_n_high_cc | +0.007 | [-0.044, +0.063] | False |
| o5a | coefs_full_model | d_clone_pct | +0.030 | [-0.066, +0.106] | False |
| o5a | coefs_coupling_model | coupling | +0.059 | [-0.118, +0.218] | False |
| o5b | coefs_level_model | erosion | +0.000 | [-0.000, +0.000] | False |
| o5b | coefs_level_model | log_n_high_cc | -0.000 | [-0.000, +0.000] | False |
| o5b | coefs_level_model | clone_pct | +0.000 | [-0.000, +0.000] | False |
| o5b | coefs_level_model | hotspot | -0.000 | [-0.000, +0.000] | False |
| o5b | coefs_full_model | erosion | +0.000 | [-0.000, +0.000] | False |
| o5b | coefs_full_model | log_n_high_cc | -0.000 | [-0.000, +0.000] | False |
| o5b | coefs_full_model | clone_pct | +0.000 | [-0.000, +0.000] | False |
| o5b | coefs_full_model | hotspot | -0.000 | [-0.000, +0.000] | False |
| o5b | coefs_full_model | d_erosion | +0.000 | [+0.000, +0.000] | True |
| o5b | coefs_full_model | d_log_n_high_cc | -0.000 | [-0.000, -0.000] | True |
| o5b | coefs_full_model | d_clone_pct | +0.000 | [-0.000, +0.000] | False |
| o5b | coefs_coupling_model | coupling | -0.000 | [-0.000, +0.000] | False |

### decision JSON
{
 "family_p_bonferroni_within": {
  "O1": 0.863568215892054,
  "O2": 0.2748625687156422,
  "O3": 0.7691154422788605,
  "O4": 0.35082458770614694,
  "O5": 0.18590704647676162
 },
 "family_p_holm": {
  "O5": 0.9295352323838081,
  "O2": 1.0,
  "O4": 1.0,
  "O3": 1.0,
  "O1": 1.0
 },
 "decisions": {
  "H1_primary_O1": {
   "level": false,
   "degradation": false
  },
  "H1_supported": false,
  "H1_partial_secondary": {
   "o2": {
    "level": false,
    "degradation": false
   },
   "o3": {
    "level": false,
    "degradation": false
   },
   "o4a": {
    "level": false,
    "degradation": false
   },
   "o4b": {
    "level": false,
    "degradation": false
   },
   "o4c": {
    "level": false,
    "degradation": false
   }
  },
  "H2_O5": {
   "o5a": {
    "level": false,
    "degradation": false
   },
   "o5b": {
    "level": false,
    "degradation": false
   }
  },
  "H2_supported": false,
  "H4_coupling_O3": false,
  "H4_supported": false,
  "construct_invalid": [
   "o5b:d_erosion(coefs_full_model)",
   "o5b:d_log_n_high_cc(coefs_full_model)"
  ]
 }
}

### file secondary rule_met
{"o1": {"level": false, "degradation": false, "coupling": false, "ownership": false}, "o2": {"level": false, "degradation": false, "coupling": false, "ownership": false}, "o3": {"level": false, "degradation": false, "coupling": false, "ownership": false}, "o5a": {"level": false, "degradation": false, "coupling": false, "ownership": false}}
