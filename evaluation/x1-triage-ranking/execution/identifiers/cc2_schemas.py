"""CC2 projection schemas derived exclusively from the named producer code."""
from identifiers.cc2_readers import ATOM, SKIP


def obj(selected, excluded=''):
    return dict({key: SKIP for key in excluded.split()}, **selected)


DATA = obj({'repo': ATOM, 'number': ATOM, 'S': ATOM, 'trail': ATOM,
    'files': [obj({'p': ATOM}, 'nloc hi nh stable nloc_t hi_t nh_t')],
    'partners': [obj({'g': ATOM, 'via': ATOM, 'vias': [ATOM], 'via_ref': ATOM}, 'n conf explained')]},
    'created merged author assoc title labels n_commits n_files additions deletions n_reviews skip '
    'S_src S_ts churn elig_churn n_elig n_new n_renamed n_old n_measured n_stable '
    'n_partner_abs n_hidden author_prior_commits')
RAW = obj({'number': ATOM, 'base_oid': ATOM, 'head_oid': ATOM, 'merge_commit': ATOM,
    'merge_parents': [ATOM], 'commits': [obj({'oid': ATOM, 'parent': ATOM}, 'authored committed email')],
    'files': [obj({'path': ATOM}, 'add dele ct')]},
    'title body labels created merged assoc additions deletions changed_files author author_type '
    'n_commits n_files n_reviews reviews')
CANDIDATE = obj({'repo': ATOM, 'parent': ATOM, 'follow': ATOM},
                'gap_h kind mention censored shared')
TYPE_CALL = obj({'repo': ATOM, 'number': ATOM},
                'rep qhash ok model_resolved choice probs http error')
PAIR_CALL = obj({'repo': ATOM, 'parent': ATOM, 'follow': ATOM},
                'rep qhash ok model_resolved choice p_corrective confidence http error')
TYPE_SAMPLE = obj({'repo': ATOM, 'number': ATOM, 'files': [ATOM]},
                  'id pipeline title labels description')
TYPE_BLIND = obj({'files': [ATOM]}, 'id title labels description')
VIEW = obj({'files': [ATOM]}, 'title description')
PAIR_SAMPLE = obj({'repo': ATOM, 'parent': ATOM, 'follow': ATOM, 'earlier': VIEW, 'later': VIEW},
                  'id stratum kind hours_between')
PAIR_BLIND = obj({'earlier': VIEW, 'later': VIEW}, 'id hours_between')
DEP_SAMPLE = obj({'repo': ATOM, 'file_a': ATOM, 'file_b': ATOM},
                 'id lang proxy partner_key head_a head_b')
DEP_BLIND = obj({'file_a': ATOM, 'file_b': ATOM}, 'id head_a head_b')
PANEL_COLUMNS = '''repo number merged created merged_ts quarter author task main_path maintainer
log_churn log_elig_files frac_new log_nloc log_author_prior Q n_measured D n_stable Q_stable
log_hidden n_hidden n_partner_abs log_partner_abs active_hours y_active n_commits y_active_g2
y_active_g8 has_review review_latency_h author_wait_h n_reviews n_changes_requested n_rework_commits
y_review y_rework censored_followup neg_duration S_src area'''.split()
SCHEMAS = {'data': DATA, 'raw': RAW, 'candidate': CANDIDATE, 'type-call': TYPE_CALL,
    'pair-call': PAIR_CALL, 'type-sample': TYPE_SAMPLE, 'type-blind': TYPE_BLIND,
    'pair-sample': PAIR_SAMPLE, 'pair-blind': PAIR_BLIND,
    'dep-sample': DEP_SAMPLE, 'dep-blind': DEP_BLIND}
# (source-relative path, schema, framing, fixed repository alias or None)
TABLES = (
    ('data/pr-django.jsonl.gz', 'data', 'jsonl', 'django'),
    ('data/pr-prometheus.jsonl.gz', 'data', 'jsonl', 'prometheus'),
    ('raw/prs-django.json.gz', 'raw', 'json', 'django'),
    ('raw/prs-prometheus.json.gz', 'raw', 'json', 'prometheus'),
    ('panel.csv.gz', 'panel', 'csv', None),
    ('panel-final.csv.gz', 'panel-final', 'csv', None),
    ('candidates.jsonl.gz', 'candidate', 'jsonl', None),
    ('types.jsonl', 'type-call', 'jsonl', None),
    ('labels.jsonl', 'pair-call', 'jsonl', None),
    ('validation/type-sample.jsonl', 'type-sample', 'jsonl', None),
    ('validation/type-blinded.jsonl', 'type-blind', 'jsonl', None),
    ('validation/pair-sample.jsonl', 'pair-sample', 'jsonl', None),
    ('validation/pair-blinded.jsonl', 'pair-blind', 'jsonl', None),
    ('validation/dep-sample.jsonl', 'dep-sample', 'jsonl', None),
    ('validation/dep-blinded.jsonl', 'dep-blind', 'jsonl', None))
PRODUCERS = ('collect.py', 'fetch_prs.py', 'common.py', 'build_panel.py', 'finalize_panel.py',
             'follow.py', 'jev_type.py', 'jev_label.py', 'validate.py', 'annotate.py',
             'constructs.py', 'analysis.py', 'gates.py', 'power.py', 'sensitivity.py', 'harness.py')
ALIASES = ('fastapi', 'webpack', 'nest', 'caddy', 'axum', 'junit5', 'eShopOnWeb', 'django',
           'pandas', 'scikit-learn', 'vite', 'fastify', 'prometheus', 'terraform', 'jellyfin', 'tokio')
TARGETS = {'django': 'django/django', 'prometheus': 'prometheus/prometheus'}
BASE = 'evaluation/change-cost-v2/'
