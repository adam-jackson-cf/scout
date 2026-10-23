# Review sheet: agent first-pass labels

Task: for each case below, read the excerpt and decide the label yourself. Confirm the proposed label or change it.
Labels are blind: the agents saw only the case state and the guide, never any tool verdict or earlier label. Every case needs your own decision; there is no bulk accept.
Guide (`triage-v1_choice/guide-1`): [GUIDE.md](GUIDE.md). Labels: real_defect | false_positive | cannot_tell.
After review, append your decisions to `labels.jsonl` as provenance `human`, by `user`.

Summary: 52 cases; proposed cannot_tell=1, false_positive=23, real_defect=28; A/B disagreements adjudicated by C: 4; unparsable: 0.
Disagreements are listed first, then the rest by case_id.

---

## 0ccaaf0222bf (A/B disagree)

- tool_signal: quality.ruff.complexity rule C901 at fastapi/encoders.py:29
- path: `fastapi/encoders.py`

`fastapi/encoders.py` lines 4-54
```
4: from pathlib import PurePath
5: from types import GeneratorType
6: from typing import Any, Callable, Dict, List, Optional, Set, Tuple, Union
7: 
8: from pydantic import BaseModel
9: from pydantic.json import ENCODERS_BY_TYPE
10: 
11: SetIntStr = Set[Union[int, str]]
12: DictIntStrAny = Dict[Union[int, str], Any]
13: 
14: 
15: def generate_encoders_by_class_tuples(
16:     type_encoder_map: Dict[Any, Callable[[Any], Any]]
17: ) -> Dict[Callable[[Any], Any], Tuple[Any, ...]]:
18:     encoders_by_class_tuples: Dict[Callable[[Any], Any], Tuple[Any, ...]] = defaultdict(
19:         tuple
20:     )
21:     for type_, encoder in type_encoder_map.items():
22:         encoders_by_class_tuples[encoder] += (type_,)
23:     return encoders_by_class_tuples
24: 
25: 
26: encoders_by_class_tuples = generate_encoders_by_class_tuples(ENCODERS_BY_TYPE)
27: 
28: 
29: def jsonable_encoder(
30:     obj: Any,
31:     include: Optional[Union[SetIntStr, DictIntStrAny]] = None,
32:     exclude: Optional[Union[SetIntStr, DictIntStrAny]] = None,
33:     by_alias: bool = True,
34:     exclude_unset: bool = False,
35:     exclude_defaults: bool = False,
36:     exclude_none: bool = False,
37:     custom_encoder: Optional[Dict[Any, Callable[[Any], Any]]] = None,
38:     sqlalchemy_safe: bool = True,
39: ) -> Any:
40:     custom_encoder = custom_encoder or {}
41:     if custom_encoder:
42:         if type(obj) in custom_encoder:
43:             return custom_encoder[type(obj)](obj)
44:         else:
45:             for encoder_type, encoder_instance in custom_encoder.items():
46:                 if isinstance(obj, encoder_type):
47:                     return encoder_instance(obj)
48:     if include is not None and not isinstance(include, (set, dict)):
49:         include = set(include)
50:     if exclude is not None and not isinstance(exclude, (set, dict)):
51:         exclude = set(exclude)
52:     if isinstance(obj, BaseModel):
53:         encoder = getattr(obj.__config__, "json_encoders", {})
54:         if custom_encoder:
```

- **A** (claude-opus-5-5): `real_defect` — rule: R5 Metrics: the excerpt shows genuinely hand-written logic that drives the metric — The flagged function `jsonable_encoder` (line 29) is shipped, non-test code. Its visible body is hand-written branching (nested if/else, a for-loop with isinstance dispatch, and several conditional normalisations), not repetitive declarative content, so the complexity metric reflects real logic.
- **B** (gpt-6.1-sol): `cannot_tell` — rule: R5 Metrics; R6 Outside the excerpt — The excerpt shows only the beginning of jsonable_encoder and some hand-written branching. The remaining function is absent, so the excerpt is too small to judge the reported complexity.
- **C** (claude-sonnet-5-5): `real_defect` — rule: R5 Metrics: genuinely hand-written logic that drives the metric — The excerpt shows shipped, hand-written code for jsonable_encoder with nested if/else, a for loop, and multiple conditional branches, so the C901 complexity reflects real branching logic rather than repetitive declarative content. R5 doesn't require the whole function to be visible; the visible branching already shows genuine logic driving the metric, and no false_positive rule applies. — agrees with: A

**Proposed label: `real_defect`**

Your decision: [ ] confirm  [ ] change to ______  note: ____

---

## bb9b50161684 (A/B disagree)

- tool_signal: quality.ruff.complexity rule C901 at .github/actions/people/app/main.py:525
- path: `.github/actions/people/app/main.py`

`.github/actions/people/app/main.py` lines 500-550
```
500: 
501: 
502: def get_experts(settings: Settings):
503:     # Migrated to only use GitHub Discussions
504:     # (
505:     #     issues_commentors,
506:     #     issues_last_month_commentors,
507:     #     issues_authors,
508:     # ) = get_issues_experts(settings=settings)
509:     (
510:         discussions_commentors,
511:         discussions_last_month_commentors,
512:         discussions_authors,
513:     ) = get_discussions_experts(settings=settings)
514:     # commentors = issues_commentors + discussions_commentors
515:     commentors = discussions_commentors
516:     # last_month_commentors = (
517:     #     issues_last_month_commentors + discussions_last_month_commentors
518:     # )
519:     last_month_commentors = discussions_last_month_commentors
520:     # authors = {**issues_authors, **discussions_authors}
521:     authors = {**discussions_authors}
522:     return commentors, last_month_commentors, authors
523: 
524: 
525: def get_contributors(settings: Settings):
526:     pr_nodes: List[PullRequestNode] = []
527:     pr_edges = get_graphql_pr_edges(settings=settings)
528: 
529:     while pr_edges:
530:         for edge in pr_edges:
531:             pr_nodes.append(edge.node)
532:         last_edge = pr_edges[-1]
533:         pr_edges = get_graphql_pr_edges(settings=settings, after=last_edge.cursor)
534: 
535:     contributors = Counter()
536:     commentors = Counter()
537:     reviewers = Counter()
538:     authors: Dict[str, Author] = {}
539: 
540:     for pr in pr_nodes:
541:         author_name = None
542:         if pr.author:
543:             authors[pr.author.login] = pr.author
544:             author_name = pr.author.login
545:         pr_commentors: Set[str] = set()
546:         pr_reviewers: Set[str] = set()
547:         for comment in pr.comments.nodes:
548:             if comment.author:
549:                 authors[comment.author.login] = comment.author
550:                 if comment.author.login == author_name:
```

- **A** (claude-opus-5-5): `real_defect` — rule: R5 Metrics: the excerpt shows genuinely hand-written logic that drives the metric — The flagged function `get_contributors` at line 525 is hand-written logic, not repetitive declarative content: a while loop, nested for loops, and several nested conditionals, still continuing past the excerpt. The `.github/actions` path is CI tooling, not a test, fixture, example, docs or generated file, so R3 does not apply.
- **B** (gpt-6.1-sol): `cannot_tell` — rule: R5 Metrics; R6 Outside the excerpt — The excerpt shows only the beginning of get_contributors and ends inside nested logic. The complete function is needed to judge whether the reported complexity reflects genuine maintainability cost.
- **C** (claude-sonnet-5-5): `real_defect` — rule: R5 Metrics: genuinely hand-written logic drives the metric — The visible portion of get_contributors shows a pagination while loop, nested for loops and multiple nested conditionals, all hand-written logic rather than repetitive declarative content. The path is shipped CI tooling, not test/fixture/generated, and nothing shows the metric is inflated, so no false_positive rule applies. — agrees with: A

**Proposed label: `real_defect`**

Your decision: [ ] confirm  [ ] change to ______  note: ____

---

## d5c3307d5189 (A/B disagree)

- tool_signal: security.gitleaks.default-credentials rule generic-api-key at src/Web/key-768c1632-cf7b-41a9-bb7a-bff228ae8fba.xml:2
- path: `src/Web/key-768c1632-cf7b-41a9-bb7a-bff228ae8fba.xml`

`src/Web/key-768c1632-cf7b-41a9-bb7a-bff228ae8fba.xml` lines 1-16
```
1: ﻿<?xml version="1.0" encoding="utf-8"?>
2: <key id="768c1632-cf7b-41a9-bb7a-bff228ae8fba" version="1">
3:   <creationDate>2021-12-01T14:37:52.0438755Z</creationDate>
4:   <activationDate>2021-12-01T14:37:52.0246578Z</activationDate>
5:   <expirationDate>2022-03-01T14:37:52.0246578Z</expirationDate>
6:   <descriptor deserializerType="Microsoft.AspNetCore.DataProtection.AuthenticatedEncryption.ConfigurationModel.AuthenticatedEncryptorDescriptorDeserializer, Microsoft.AspNetCore.DataProtection, Version=6.0.0.0, Culture=neutral, PublicKeyToken=adb9793829ddae60">
7:     <descriptor>
8:       <encryption algorithm="AES_256_CBC" />
9:       <validation algorithm="HMACSHA256" />
10:       <masterKey p4:requiresEncryption="true" xmlns:p4="http://schemas.asp.net/2015/03/dataProtection">
11:         <!-- Warning: the key below is in an unencrypted form. -->
12:         <value>PF3GdfO7PnvHYvXyD5nxmoQ91pY9qfA0rjRsdXHdUQbE1Mg9Xok2gXLY2zn8XemsySH37UGrGknht8u/PlehWg==</value>
13:       </masterKey>
14:     </descriptor>
15:   </descriptor>
16: </key>
```

- **A** (claude-opus-5-5): `real_defect` — rule: R1 Presence (R3 not applicable: `src/Web/` is a shipped path, not a test, fixture, example, docs or generated-file path) — Line 12 contains a literal AES-256/HMAC master key for ASP.NET Core Data Protection, and line 11 explicitly warns that it is stored unencrypted. The file sits in shipped source under `src/Web/`, and nothing in the excerpt guards or neutralises the exposed key.
- **B** (gpt-6.1-sol): `false_positive` — rule: R1 Presence — The flagged line contains a UUID identifying an XML key record, not an API key or default credential. The master key elsewhere in the excerpt does not establish the reported generic-api-key mechanism.
- **C** (claude-sonnet-5-5): `real_defect` — rule: R1 Presence — The file is shipped source under src/Web/ and line 12 contains a literal unencrypted ASP.NET Core Data Protection master key (line 11 warns of this), with no guard or neutralising construct visible. The flagged line 2 is only a key ID, but the same file's excerpt shows the actual secret literally present, so the flagged mechanism (a hardcoded key) is present. — agrees with: A

**Proposed label: `real_defect`**

Your decision: [ ] confirm  [ ] change to ______  note: ____

---

## typescript-quality-route-params-complexity-676c98ca (A/B disagree)

- tool_signal: lizard: cyclomatic complexity 21 at line 9 of packages/core/router/route-params-factory.ts (high complexity)
- path: `packages/core/router/route-params-factory.ts`

`packages/core/router/route-params-factory.ts` lines 1-45
```
1: import { RouteParamtypes } from '@nestjs/common/enums/route-paramtypes.enum';
2: import { IRouteParamsFactory } from './interfaces/route-params-factory.interface';
3: 
4: export class RouteParamsFactory implements IRouteParamsFactory {
5:   public exchangeKeyForValue<
6:     TRequest extends Record<string, any> = any,
7:     TResponse = any,
8:     TResult = any,
9:   >(
10:     key: RouteParamtypes | string,
11:     data: string | object | any,
12:     { req, res, next }: { req: TRequest; res: TResponse; next: Function },
13:   ): TResult {
14:     switch (key) {
15:       case RouteParamtypes.NEXT:
16:         return next as any;
17:       case RouteParamtypes.REQUEST:
18:         return req as any;
19:       case RouteParamtypes.RESPONSE:
20:         return res as any;
21:       case RouteParamtypes.BODY:
22:         return data && req.body ? req.body[data] : req.body;
23:       case RouteParamtypes.PARAM:
24:         return data ? req.params[data] : req.params;
25:       case RouteParamtypes.HOST:
26:         const hosts = req.hosts || {};
27:         return data ? hosts[data] : hosts;
28:       case RouteParamtypes.QUERY:
29:         return data ? req.query[data] : req.query;
30:       case RouteParamtypes.HEADERS:
31:         return data ? req.headers[data.toLowerCase()] : req.headers;
32:       case RouteParamtypes.SESSION:
33:         return req.session;
34:       case RouteParamtypes.FILE:
35:         return req[data || 'file'];
36:       case RouteParamtypes.FILES:
37:         return req.files;
38:       case RouteParamtypes.IP:
39:         return req.ip;
40:       default:
41:         return null;
42:     }
43:   }
44: }
45: 
```

- **A** (claude-opus-5-5): `false_positive` — rule: R5 Metrics — The complexity of 21 comes almost entirely from a flat switch that maps each RouteParamtypes key to a one-line request accessor, in effect a lookup table. This repetitive dispatch inflates the metric without reflecting real maintainability cost.
- **B** (gpt-6.1-sol): `real_defect` — rule: R5 Metrics — The full method shows hand-written dispatch logic with multiple cases, conditional expressions, and fallback operations that drive the complexity metric. It is shipped code, and the excerpt does not show repetitive declarative content inflating the metric.
- **C** (claude-sonnet-5-5): `false_positive` — rule: R5 Metrics — The full method is visible and is a flat switch where each case is a one-line accessor mapping a RouteParamtypes key to a request property, so the cyclomatic count of 21 is inflated by repetitive dispatch rather than by intricate logic. The few ternaries are trivial 'data ? x[data] : x' patterns, so the metric does not reflect real maintainability cost. — agrees with: A

**Proposed label: `false_positive`**

Your decision: [ ] confirm  [ ] change to ______  note: ____

---

## 0ff5a484e29c

- tool_signal: security.semgrep.actions-mutable-ref rule scout.security.actions.mutable-ref at .github/workflows/codeql-analysis.yml:27
- path: `.github/workflows/codeql-analysis.yml`

`.github/workflows/codeql-analysis.yml` lines 2-39
```
2: 
3: on:
4:   push:
5:     branches: [main, releases/**]
6:   pull_request:
7:     # The branches below must be a subset of the branches above
8:     branches: [main, releases/**]
9:   schedule:
10:     - cron: '0 19 * * 3'
11: 
12: env:
13:   GRADLE_ENTERPRISE_ACCESS_KEY: ${{ secrets.GRADLE_ENTERPRISE_ACCESS_KEY }}
14: 
15: jobs:
16:   analyze:
17:     name: Analyze
18:     runs-on: ubuntu-latest
19:     strategy:
20:       fail-fast: false
21:       matrix:
22:         language: ['java', 'javascript']
23:     steps:
24:     - name: Check out repository
25:       uses: actions/checkout@v3
26:     - name: Initialize CodeQL
27:       uses: github/codeql-action/init@v1
28:       with:
29:         languages: ${{ matrix.language }}
30:     - name: Build
31:       uses: ./.github/actions/run-gradle
32:       with:
33:         arguments: |
34:           --no-build-cache
35:           -Dscan.tag.CodeQL
36:           allMainClasses
37:     - name: Perform CodeQL Analysis
38:       uses: github/codeql-action/analyze@v1
39: 
```

- **A** (claude-opus-5-5): `real_defect` — rule: R1 Presence (no false_positive rule R2–R5 applies) — Line 27 uses `github/codeql-action/init@v1`, a mutable major-version tag rather than a commit SHA pin, in a real CI workflow (not test/fixture/docs). Nothing in the excerpt pins or neutralises it.
- **B** (gpt-6.1-sol): `real_defect` — rule: R1 Presence — Line 27 uses github/codeql-action/init@v1, a mutable tag rather than an immutable commit pin. The workflow excerpt shows no guard or neutralising construct.

**Proposed label: `real_defect`**

Your decision: [ ] confirm  [ ] change to ______  note: ____

---

## 1154b6cdfcfc

- tool_signal: security.semgrep.actions-mutable-ref rule scout.security.actions.mutable-ref at .github/workflows/test.yml:39
- path: `.github/workflows/test.yml`

`.github/workflows/test.yml` lines 14-64
```
14:       - dev-1
15: 
16: permissions:
17:   contents: read
18: 
19: jobs:
20:   lint:
21:     runs-on: ubuntu-latest
22:     steps:
23:       - uses: actions/checkout@v3
24:       - name: Use Node.js
25:         uses: actions/setup-node@v3
26:         with:
27:           node-version: 16.x
28:           cache: "yarn"
29:       - run: yarn --frozen-lockfile
30:       - uses: actions/cache@v3
31:         with:
32:           path: .eslintcache
33:           key: lint-${{ env.GITHUB_SHA }}
34:           restore-keys: lint-
35:       - run: yarn lint
36:   basic:
37:     runs-on: ubuntu-latest
38:     steps:
39:       - uses: actions/checkout@v3
40:       - name: Use Node.js
41:         uses: actions/setup-node@v3
42:         with:
43:           node-version: 16.x
44:           cache: "yarn"
45:       - run: yarn --frozen-lockfile
46:       - run: yarn link --frozen-lockfile || true
47:       - run: yarn link webpack --frozen-lockfile
48:       - run: yarn test:basic --ci
49:       - uses: codecov/codecov-action@v3
50:         with:
51:           flags: basic
52:           functionalities: gcov
53:   validate-legacy-node:
54:     runs-on: ubuntu-latest
55:     steps:
56:       - uses: actions/checkout@v3
57:       - name: Use Node.js
58:         uses: actions/setup-node@v3
59:         with:
60:           node-version: 10.x
61:           cache: "yarn"
62:       # Remove `devDependencies` from `package.json` to avoid `yarn install` compatibility error
63:       - run: node -e "const content = require('./package.json');delete content.devDependencies;require('fs').writeFileSync('package.json', JSON.stringify(content, null, 2));"
64:       - run: yarn install --production --frozen-lockfile
```

- **A** (claude-opus-5-5): `real_defect` — rule: R1 Presence / R2 unguarded (R3 does not apply) — Line 39 literally uses `actions/checkout@v3`, a mutable tag rather than a commit SHA, and nothing in the excerpt pins or neutralises it. The file is a hand-written CI workflow that only runs tests; it is not a test or fixture file, so R3 does not apply.
- **B** (gpt-6.1-sol): `real_defect` — rule: R1 Presence — Line 39 uses actions/checkout@v3, a mutable tag rather than an immutable commit pin, with no visible construct neutralising it. The workflow filename test.yml does not make this CI configuration a test or fixture file.

**Proposed label: `real_defect`**

Your decision: [ ] confirm  [ ] change to ______  note: ____

---

## 140801a4a2dc

- tool_signal: security.semgrep.actions-mutable-ref rule scout.security.actions.mutable-ref at .github/workflows/richnav.yml:13
- path: `.github/workflows/richnav.yml`

`.github/workflows/richnav.yml` lines 1-25
```
1: name: eShopOnWeb - Code Index
2: 
3: on: workflow_dispatch
4: 
5: jobs:
6:   build:
7: 
8:     runs-on: windows-latest
9: 
10:     steps:
11:     - uses: actions/checkout@v2
12:     - name: Setup .NET Core
13:       uses: actions/setup-dotnet@v1
14:       with:
15:         dotnet-version: 8.0.x
16: 
17:     - name: Build with dotnet
18:       run: dotnet build ./Everything.sln --configuration Release /bl
19: 
20:     - uses: microsoft/RichCodeNavIndexer@v0.1
21:       with:
22:         repo-token: ${{ github.token }}
23:         languages: 'csharp'
24:         environment: 'internal'
25: 
```

- **A** (claude-opus-5-5): `real_defect` — rule: R1 Presence / R2 no guard — Line 13 literally uses `actions/setup-dotnet@v1`, a mutable tag rather than a pinned commit SHA, in a shipped workflow file. Nothing in the excerpt pins or otherwise neutralises the mutable reference.
- **B** (gpt-6.1-sol): `real_defect` — rule: R1 Presence — Line 13 uses actions/setup-dotnet@v1, a mutable tag rather than an immutable commit pin. The workflow excerpt shows no guard or neutralising construct.

**Proposed label: `real_defect`**

Your decision: [ ] confirm  [ ] change to ______  note: ____

---

## 17871ab8c5fd

- tool_signal: security.semgrep.actions-mutable-ref rule scout.security.actions.mutable-ref at .github/workflows/richnav.yml:11
- path: `.github/workflows/richnav.yml`

`.github/workflows/richnav.yml` lines 1-25
```
1: name: eShopOnWeb - Code Index
2: 
3: on: workflow_dispatch
4: 
5: jobs:
6:   build:
7: 
8:     runs-on: windows-latest
9: 
10:     steps:
11:     - uses: actions/checkout@v2
12:     - name: Setup .NET Core
13:       uses: actions/setup-dotnet@v1
14:       with:
15:         dotnet-version: 8.0.x
16: 
17:     - name: Build with dotnet
18:       run: dotnet build ./Everything.sln --configuration Release /bl
19: 
20:     - uses: microsoft/RichCodeNavIndexer@v0.1
21:       with:
22:         repo-token: ${{ github.token }}
23:         languages: 'csharp'
24:         environment: 'internal'
25: 
```

- **A** (claude-opus-5-5): `real_defect` — rule: R1 Presence (mechanism literally present, unguarded, shipped config) — Line 11 uses `actions/checkout@v2`, a mutable tag rather than a pinned commit SHA, in a real CI workflow file with no visible pinning or guard. The .github/workflows path is shipped repository configuration, not a test/fixture/example path.
- **B** (gpt-6.1-sol): `real_defect` — rule: R1 Presence — Line 11 uses actions/checkout@v2, a mutable tag rather than an immutable commit pin. The workflow excerpt shows no guard or neutralising construct.

**Proposed label: `real_defect`**

Your decision: [ ] confirm  [ ] change to ______  note: ____

---

## 1b1f5f714a4e

- tool_signal: security.semgrep.actions-mutable-ref rule scout.security.actions.mutable-ref at .github/workflows/cross-version.yml:27
- path: `.github/workflows/cross-version.yml`

`.github/workflows/cross-version.yml` lines 2-52
```
2: 
3: on:
4:   push:
5:     branches:
6:       - main
7:       - 'releases/*'
8:   pull_request:
9:     branches:
10:       - '*'
11: 
12: env:
13:   ORG_GRADLE_PROJECT_enableTestDistribution: true
14:   ORG_GRADLE_PROJECT_junitBuildCacheUsername: ${{ secrets.BUILD_CACHE_USERNAME }}
15:   ORG_GRADLE_PROJECT_junitBuildCachePassword: ${{ secrets.BUILD_CACHE_PASSWORD }}
16:   GRADLE_ENTERPRISE_ACCESS_KEY: ${{ secrets.GRADLE_ENTERPRISE_ACCESS_KEY }}
17: 
18: jobs:
19:   openjdk:
20:     strategy:
21:       matrix:
22:         jdk: [19, 20, 21]
23:     name: "OpenJDK ${{ matrix.jdk }}"
24:     runs-on: ubuntu-latest
25:     steps:
26:       - name: Check out repository
27:         uses: actions/checkout@v3
28:         with:
29:           fetch-depth: 1
30:       - name: Set up Test JDK
31:         uses: ./.github/actions/setup-test-jdk
32:       - name: 'Set up JDK ${{ matrix.jdk }}'
33:         uses: oracle-actions/setup-java@v1
34:         with:
35:           website: jdk.java.net
36:           release: ${{ matrix.jdk }}
37:           version: latest
38:       - name: 'Prepare JDK${{ matrix.jdk }} env var'
39:         shell: bash
40:         run: echo "JDK${{ matrix.jdk }}=$JAVA_HOME" >> $GITHUB_ENV
41:       - name: Build
42:         uses: ./.github/actions/run-gradle
43:         with:
44:           arguments: |
45:             -PjavaToolchainVersion=${{ matrix.jdk }}
46:             -Dscan.tag.JDK_${{ matrix.jdk }}
47:             build
48:       - name: Upload Test Distribution trace files
49:         uses: actions/upload-artifact@v3
50:         with:
51:           name: "Test Distribution trace files (OpenJDK ${{ matrix.jdk }})"
52:           path: '**/build/test-results/*/trace.json'
```

- **A** (claude-opus-5-5): `real_defect` — rule: R1 Presence / R2 unguarded — Line 27 literally uses `actions/checkout@v3`, a mutable tag rather than a pinned commit SHA, in a shipped CI workflow. Nothing in the excerpt pins or otherwise neutralises the reference.
- **B** (gpt-6.1-sol): `real_defect` — rule: R1 Presence — Line 27 uses actions/checkout@v3, a mutable tag rather than an immutable commit pin. The workflow shows no guard or neutralising construct addressing that mutable reference.

**Proposed label: `real_defect`**

Your decision: [ ] confirm  [ ] change to ______  note: ____

---

## 3c90020b1500

- tool_signal: quality.ruff.complexity rule C901 at fastapi/openapi/utils.py:393
- path: `fastapi/openapi/utils.py`

`fastapi/openapi/utils.py` lines 368-418
```
368:     for route in routes:
369:         if getattr(route, "include_in_schema", None) and isinstance(
370:             route, routing.APIRoute
371:         ):
372:             if route.body_field:
373:                 assert isinstance(
374:                     route.body_field, ModelField
375:                 ), "A request body must be a Pydantic Field"
376:                 body_fields_from_routes.append(route.body_field)
377:             if route.response_field:
378:                 responses_from_routes.append(route.response_field)
379:             if route.response_fields:
380:                 responses_from_routes.extend(route.response_fields.values())
381:             if route.callbacks:
382:                 callback_flat_models |= get_flat_models_from_routes(route.callbacks)
383:             params = get_flat_params(route.dependant)
384:             request_fields_from_routes.extend(params)
385: 
386:     flat_models = callback_flat_models | get_flat_models_from_fields(
387:         body_fields_from_routes + responses_from_routes + request_fields_from_routes,
388:         known_models=set(),
389:     )
390:     return flat_models
391: 
392: 
393: def get_openapi(
394:     *,
395:     title: str,
396:     version: str,
397:     openapi_version: str = "3.0.2",
398:     description: Optional[str] = None,
399:     routes: Sequence[BaseRoute],
400:     tags: Optional[List[Dict[str, Any]]] = None,
401:     servers: Optional[List[Dict[str, Union[str, Any]]]] = None,
402:     terms_of_service: Optional[str] = None,
403:     contact: Optional[Dict[str, Union[str, Any]]] = None,
404:     license_info: Optional[Dict[str, Union[str, Any]]] = None,
405: ) -> Dict[str, Any]:
406:     info: Dict[str, Any] = {"title": title, "version": version}
407:     if description:
408:         info["description"] = description
409:     if terms_of_service:
410:         info["termsOfService"] = terms_of_service
411:     if contact:
412:         info["contact"] = contact
413:     if license_info:
414:         info["license"] = license_info
415:     output: Dict[str, Any] = {"openapi": openapi_version, "info": info}
416:     if servers:
417:         output["servers"] = servers
418:     components: Dict[str, Dict[str, Any]] = {}
```

- **A** (claude-opus-5-5): `cannot_tell` — rule: R5 (excerpt too small to judge) / R6 (outside the excerpt) — The excerpt shows only the first ~25 lines of `get_openapi`. Those lines are repetitive optional-field `if` assignments, but the rest of the function body, which likely drives the C901 score, is cut off at line 418. So it can't be judged whether the complexity is declarative inflation or genuine hand-written logic.
- **B** (gpt-6.1-sol): `cannot_tell` — rule: R5, R6 — The excerpt shows only the beginning of get_openapi, so its overall complexity cannot be assessed. The remaining function body is outside the excerpt.

**Proposed label: `cannot_tell`**

Your decision: [ ] confirm  [ ] change to ______  note: ____

---

## 404abe793491

- tool_signal: security.semgrep.actions-mutable-ref rule scout.security.actions.mutable-ref at .github/workflows/test.yml:56
- path: `.github/workflows/test.yml`

`.github/workflows/test.yml` lines 31-81
```
31:         with:
32:           path: .eslintcache
33:           key: lint-${{ env.GITHUB_SHA }}
34:           restore-keys: lint-
35:       - run: yarn lint
36:   basic:
37:     runs-on: ubuntu-latest
38:     steps:
39:       - uses: actions/checkout@v3
40:       - name: Use Node.js
41:         uses: actions/setup-node@v3
42:         with:
43:           node-version: 16.x
44:           cache: "yarn"
45:       - run: yarn --frozen-lockfile
46:       - run: yarn link --frozen-lockfile || true
47:       - run: yarn link webpack --frozen-lockfile
48:       - run: yarn test:basic --ci
49:       - uses: codecov/codecov-action@v3
50:         with:
51:           flags: basic
52:           functionalities: gcov
53:   validate-legacy-node:
54:     runs-on: ubuntu-latest
55:     steps:
56:       - uses: actions/checkout@v3
57:       - name: Use Node.js
58:         uses: actions/setup-node@v3
59:         with:
60:           node-version: 10.x
61:           cache: "yarn"
62:       # Remove `devDependencies` from `package.json` to avoid `yarn install` compatibility error
63:       - run: node -e "const content = require('./package.json');delete content.devDependencies;require('fs').writeFileSync('package.json', JSON.stringify(content, null, 2));"
64:       - run: yarn install --production --frozen-lockfile
65:   unit:
66:     runs-on: ubuntu-latest
67:     steps:
68:       - uses: actions/checkout@v3
69:       - name: Use Node.js
70:         uses: actions/setup-node@v3
71:         with:
72:           node-version: 16.x
73:           cache: "yarn"
74:       - run: yarn --frozen-lockfile
75:       - run: yarn link --frozen-lockfile || true
76:       - run: yarn link webpack --frozen-lockfile
77:       - uses: actions/cache@v3
78:         with:
79:           path: .jest-cache
80:           key: jest-unit-${{ env.GITHUB_SHA }}
81:           restore-keys: jest-unit-
```

- **A** (claude-opus-5-5): `real_defect` — rule: R1 Presence (R3 does not apply) — Line 56 uses `actions/checkout@v3`, a mutable tag rather than a pinned commit SHA, and nothing in the excerpt pins or guards it. The file is a hand-written CI workflow; its name `test.yml` only resembles a test file and is not test or fixture code.
- **B** (gpt-6.1-sol): `real_defect` — rule: R1 Presence — Line 56 uses actions/checkout@v3, a mutable tag, with no visible immutable commit pin or neutralising guard. The file is an executable GitHub Actions workflow; its test-oriented name does not make the action reference test or fixture code.

**Proposed label: `real_defect`**

Your decision: [ ] confirm  [ ] change to ______  note: ____

---

## 4e97ec3b7db0

- tool_signal: security.gitleaks.default-credentials rule generic-api-key at junit-jupiter-engine/src/test/java/org/junit/jupiter/api/condition/DisabledIfEnvironmentVariableIntegrationTests.java:38
- path: `junit-jupiter-engine/src/test/java/org/junit/jupiter/api/condition/DisabledIfEnvironmentVariableIntegrationTests.java`

`junit-jupiter-engine/src/test/java/org/junit/jupiter/api/condition/DisabledIfEnvironmentVariableIntegrationTests.java` lines 13-63
```
13: import static org.junit.jupiter.api.Assertions.assertNotEquals;
14: import static org.junit.jupiter.api.Assertions.assertNull;
15: import static org.junit.jupiter.api.Assertions.fail;
16: 
17: import java.lang.annotation.ElementType;
18: import java.lang.annotation.Retention;
19: import java.lang.annotation.RetentionPolicy;
20: import java.lang.annotation.Target;
21: 
22: import org.junit.jupiter.api.Disabled;
23: import org.junit.jupiter.api.Test;
24: 
25: /**
26:  * Integration tests for {@link DisabledIfEnvironmentVariable}.
27:  *
28:  * @since 5.1
29:  */
30: @Disabled("Disabled since the required environment variables are not set")
31: // Tests (except those with intentional configuration errors) will pass if you set
32: // the following environment variables:
33: // DisabledIfEnvironmentVariableTests.key1 = enigma
34: // DisabledIfEnvironmentVariableTests.key2 = enigma
35: class DisabledIfEnvironmentVariableIntegrationTests {
36: 
37: 	static final String KEY1 = "DisabledIfEnvironmentVariableTests.key1";
38: 	static final String KEY2 = "DisabledIfEnvironmentVariableTests.key2";
39: 	static final String ENIGMA = "enigma";
40: 	static final String BOGUS = "bogus";
41: 
42: 	@Test
43: 	void enabledBecauseAnnotationIsNotPresent() {
44: 	}
45: 
46: 	@Test
47: 	@DisabledIfEnvironmentVariable(named = "  ", matches = ENIGMA)
48: 	void blankNamedAttribute() {
49: 	}
50: 
51: 	@Test
52: 	@DisabledIfEnvironmentVariable(named = KEY1, matches = "  ")
53: 	void blankMatchesAttribute() {
54: 	}
55: 
56: 	@Test
57: 	@DisabledIfEnvironmentVariable(named = KEY1, matches = ENIGMA, disabledReason = "That's an enigma")
58: 	void disabledBecauseEnvironmentVariableMatchesExactly() {
59: 		fail("should be disabled");
60: 	}
61: 
62: 	@Test
63: 	@DisabledIfEnvironmentVariable(named = KEY1, matches = BOGUS)
```

- **A** (claude-opus-5-5): `false_positive` — rule: R3 Path context (also R1: mechanism absent) — The file is under `src/test/java/` and is an integration test class. Line 38 is only an environment-variable name constant (`"DisabledIfEnvironmentVariableTests.key2"`), not a credential or API key.
- **B** (gpt-6.1-sol): `false_positive` — rule: R3 Path context — The flagged line is in src/test/java in an integration test class. It declares an environment-variable name used by tests, not an API credential.

**Proposed label: `false_positive`**

Your decision: [ ] confirm  [ ] change to ______  note: ____

---

## 6279a9cf3a91

- tool_signal: security.semgrep.actions-mutable-ref rule scout.security.actions.mutable-ref at .github/workflows/smokeshow.yml:23
- path: `.github/workflows/smokeshow.yml`

`.github/workflows/smokeshow.yml` lines 1-36
```
1: name: Smokeshow
2: 
3: on:
4:   workflow_run:
5:     workflows: [Test]
6:     types: [completed]
7: 
8: permissions:
9:   statuses: write
10: 
11: jobs:
12:   smokeshow:
13:     if: ${{ github.event.workflow_run.conclusion == 'success' }}
14:     runs-on: ubuntu-latest
15: 
16:     steps:
17:       - uses: actions/setup-python@v4
18:         with:
19:           python-version: '3.9'
20: 
21:       - run: pip install smokeshow
22: 
23:       - uses: dawidd6/action-download-artifact@v2.26.0
24:         with:
25:           workflow: test.yml
26:           commit: ${{ github.event.workflow_run.head_sha }}
27: 
28:       - run: smokeshow upload coverage-html
29:         env:
30:           SMOKESHOW_GITHUB_STATUS_DESCRIPTION: Coverage {coverage-percentage}
31:           SMOKESHOW_GITHUB_COVERAGE_THRESHOLD: 100
32:           SMOKESHOW_GITHUB_CONTEXT: coverage
33:           SMOKESHOW_GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
34:           SMOKESHOW_GITHUB_PR_HEAD_SHA: ${{ github.event.workflow_run.head_sha }}
35:           SMOKESHOW_AUTH_KEY: ${{ secrets.SMOKESHOW_AUTH_KEY }}
36: 
```

- **A** (claude-opus-5-5): `real_defect` — rule: R1 Presence — Line 23 literally uses `dawidd6/action-download-artifact@v2.26.0`, a mutable tag rather than a commit SHA, in a shipped workflow file with secrets in scope. Nothing in the excerpt pins or otherwise neutralises the reference.
- **B** (gpt-6.1-sol): `real_defect` — rule: R1 Presence — Line 23 references the action using the mutable tag v2.26.0 rather than an immutable commit SHA. No visible guard or pin neutralises that mutable reference.

**Proposed label: `real_defect`**

Your decision: [ ] confirm  [ ] change to ______  note: ____

---

## 675379ba1c07

- tool_signal: security.semgrep.actions-mutable-ref rule scout.security.actions.mutable-ref at .github/workflows/ci.yml:52
- path: `.github/workflows/ci.yml`

`.github/workflows/ci.yml` lines 27-77
```
27:         - go: '1.18'
28:           GO_SEMVER: '~1.18.4'
29: 
30:         - go: '1.20'
31:           GO_SEMVER: '~1.20.0'
32: 
33:         # Set some variables per OS, usable via ${{ matrix.VAR }}
34:         # CADDY_BIN_PATH: the path to the compiled Caddy binary, for artifact publishing
35:         # SUCCESS: the typical value for $? per OS (Windows/pwsh returns 'True')
36:         - os: ubuntu-latest
37:           CADDY_BIN_PATH: ./cmd/caddy/caddy
38:           SUCCESS: 0
39: 
40:         - os: macos-latest
41:           CADDY_BIN_PATH: ./cmd/caddy/caddy
42:           SUCCESS: 0
43: 
44:         - os: windows-latest
45:           CADDY_BIN_PATH: ./cmd/caddy/caddy.exe
46:           SUCCESS: 'True'
47: 
48:     runs-on: ${{ matrix.os }}
49: 
50:     steps:
51:     - name: Install Go
52:       uses: actions/setup-go@v3
53:       with:
54:         go-version: ${{ matrix.GO_SEMVER }}
55:         check-latest: true
56: 
57:     - name: Checkout code
58:       uses: actions/checkout@v3
59: 
60:     # These tools would be useful if we later decide to reinvestigate
61:     # publishing test/coverage reports to some tool for easier consumption
62:     # - name: Install test and coverage analysis tools
63:     #   run: |
64:     #     go get github.com/axw/gocov/gocov
65:     #     go get github.com/AlekSi/gocov-xml
66:     #     go get -u github.com/jstemmer/go-junit-report
67:     #     echo "$(go env GOPATH)/bin" >> $GITHUB_PATH
68: 
69:     - name: Print Go version and environment
70:       id: vars
71:       run: |
72:         printf "Using go at: $(which go)\n"
73:         printf "Go version: $(go version)\n"
74:         printf "\n\nGo environment:\n\n"
75:         go env
76:         printf "\n\nSystem environment:\n\n"
77:         env
```

- **A** (claude-opus-5-5): `real_defect` — rule: R1 Presence (no R2/R3 guard applies) — Line 52 uses `actions/setup-go@v3`, a mutable tag rather than a pinned commit SHA, in a shipped CI workflow. Nothing in the excerpt pins or neutralises it, and `.github/workflows/` is not a test, fixture or generated path.
- **B** (gpt-6.1-sol): `real_defect` — rule: R1 Presence — Line 52 uses actions/setup-go@v3, a mutable tag rather than an immutable commit pin. The workflow excerpt shows no guard or neutralising construct addressing the mutable reference.

**Proposed label: `real_defect`**

Your decision: [ ] confirm  [ ] change to ______  note: ____

---

## 6c731d4c311a

- tool_signal: security.semgrep.actions-mutable-ref rule scout.security.actions.mutable-ref at .github/workflows/ci.yml:178
- path: `.github/workflows/ci.yml`

`.github/workflows/ci.yml` lines 153-184
```
153:           mkdir -p ~/.ssh && echo -e "${SSH_KEY//_/\\n}" > ~/.ssh/id_ecdsa && chmod og-rwx ~/.ssh/id_ecdsa
154: 
155:           # short sha is enough?
156:           short_sha=$(git rev-parse --short HEAD)
157: 
158:           # The environment is fresh, so there's no point in keeping accepting and adding the key.
159:           rsync -arz -e "ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null" --progress --delete --exclude '.git' . "$CI_USER"@ci-s390x.caddyserver.com:/var/tmp/"$short_sha"
160:           ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -t "$CI_USER"@ci-s390x.caddyserver.com "cd /var/tmp/$short_sha; go version; go env; printf "\n\n";CGO_ENABLED=0 go test -v ./..."
161:           test_result=$?
162: 
163:           # There's no need leaving the files around
164:           ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null "$CI_USER"@ci-s390x.caddyserver.com "rm -rf /var/tmp/'$short_sha'"
165: 
166:           echo "Test exit code: $test_result"
167:           exit $test_result
168:         env:
169:           SSH_KEY: ${{ secrets.S390X_SSH_KEY }}
170:           CI_USER: ${{ secrets.CI_USER }}
171: 
172:   goreleaser-check:
173:     runs-on: ubuntu-latest
174:     steps:
175:       - name: checkout
176:         uses: actions/checkout@v3
177:       
178:       - uses: goreleaser/goreleaser-action@v4
179:         with:
180:           version: latest
181:           args: check
182:         env:
183:           TAG: ${{ steps.vars.outputs.version_tag }}
184: 
```

- **A** (claude-opus-5-5): `real_defect` — rule: R1 Presence (no R2/R3 guard applies) — Line 178 literally uses `goreleaser/goreleaser-action@v4`, a mutable tag rather than a pinned commit SHA, in the repository's own CI workflow, and nothing in the excerpt pins or neutralises it. The `.github/workflows/` path is shipped CI configuration, not a test/fixture/example path, so R3 does not apply.
- **B** (gpt-6.1-sol): `real_defect` — rule: R1 Presence — Line 178 references goreleaser/goreleaser-action@v4, a mutable tag rather than an immutable commit SHA. No visible pin or guard neutralises the mutable reference.

**Proposed label: `real_defect`**

Your decision: [ ] confirm  [ ] change to ______  note: ____

---

## 7484113ca3c6

- tool_signal: performance:ruff rule PERF401 at .github/actions/people/app/main.py:430
- path: `.github/actions/people/app/main.py`

`.github/actions/people/app/main.py` lines 405-455
```
405:         after=after,
406:         category_id=questions_category_id,
407:     )
408:     graphql_response = DiscussionsResponse.parse_obj(data)
409:     return graphql_response.data.repository.discussions.edges
410: 
411: 
412: def get_graphql_pr_edges(*, settings: Settings, after: Union[str, None] = None):
413:     data = get_graphql_response(settings=settings, query=prs_query, after=after)
414:     graphql_response = PRsResponse.parse_obj(data)
415:     return graphql_response.data.repository.pullRequests.edges
416: 
417: 
418: def get_graphql_sponsor_edges(*, settings: Settings, after: Union[str, None] = None):
419:     data = get_graphql_response(settings=settings, query=sponsors_query, after=after)
420:     graphql_response = SponsorsResponse.parse_obj(data)
421:     return graphql_response.data.user.sponsorshipsAsMaintainer.edges
422: 
423: 
424: def get_issues_experts(settings: Settings):
425:     issue_nodes: List[IssuesNode] = []
426:     issue_edges = get_graphql_issue_edges(settings=settings)
427: 
428:     while issue_edges:
429:         for edge in issue_edges:
430:             issue_nodes.append(edge.node)
431:         last_edge = issue_edges[-1]
432:         issue_edges = get_graphql_issue_edges(settings=settings, after=last_edge.cursor)
433: 
434:     commentors = Counter()
435:     last_month_commentors = Counter()
436:     authors: Dict[str, Author] = {}
437: 
438:     now = datetime.now(tz=timezone.utc)
439:     one_month_ago = now - timedelta(days=30)
440: 
441:     for issue in issue_nodes:
442:         issue_author_name = None
443:         if issue.author:
444:             authors[issue.author.login] = issue.author
445:             issue_author_name = issue.author.login
446:         issue_commentors = set()
447:         for comment in issue.comments.nodes:
448:             if comment.author:
449:                 authors[comment.author.login] = comment.author
450:                 if comment.author.login != issue_author_name:
451:                     issue_commentors.add(comment.author.login)
452:         for author_name in issue_commentors:
453:             commentors[author_name] += 1
454:             if issue.createdAt > one_month_ago:
455:                 last_month_commentors[author_name] += 1
```

- **A** (claude-opus-5-5): `real_defect` — rule: R1 Presence — Line 430 is a plain for-loop that only calls `issue_nodes.append(edge.node)`, which is exactly the pattern PERF401 flags (it could be `issue_nodes.extend(...)`), and nothing in the excerpt guards or neutralises it. The `.github/actions/` path is hand-written action code, not a test, fixture, example, docs or generated file, so R3 does not apply.
- **B** (gpt-6.1-sol): `real_defect` — rule: R1 Presence — Lines 429–430 visibly use a for loop to append each edge.node to a list, the mechanism flagged by PERF401. This is hand-written action code with no visible neutralising construct or excluded file context.

**Proposed label: `real_defect`**

Your decision: [ ] confirm  [ ] change to ______  note: ____

---

## 76f4a9301717

- tool_signal: security.semgrep.actions-mutable-ref rule scout.security.actions.mutable-ref at .github/workflows/gradle-wrapper-validation.yml:10
- path: `.github/workflows/gradle-wrapper-validation.yml`

`.github/workflows/gradle-wrapper-validation.yml` lines 1-15
```
1: name: "Validate Gradle Wrapper"
2: on: [push, pull_request]
3: 
4: jobs:
5:   validation:
6:     name: "Validation"
7:     runs-on: ubuntu-latest
8:     steps:
9:       - name: Check out repository
10:         uses: actions/checkout@v3
11:         with:
12:           fetch-depth: 1
13:       - name: Validate Gradle wrapper
14:         uses: gradle/wrapper-validation-action@v1
15: 
```

- **A** (claude-opus-5-5): `real_defect` — rule: R1 Presence — Line 10 literally uses `actions/checkout@v3`, a mutable tag rather than a pinned commit SHA, in a shipped CI workflow file. Nothing in the excerpt pins or otherwise neutralises the reference.
- **B** (gpt-6.1-sol): `real_defect` — rule: R1 Presence — Line 10 uses actions/checkout@v3, a mutable tag rather than an immutable commit pin. The workflow shows no guard or neutralising construct.

**Proposed label: `real_defect`**

Your decision: [ ] confirm  [ ] change to ______  note: ____

---

## 7742b9346f37

- tool_signal: security.gitleaks.default-credentials rule generic-api-key at junit-jupiter-engine/src/test/java/org/junit/jupiter/api/condition/EnabledIfEnvironmentVariableIntegrationTests.java:36
- path: `junit-jupiter-engine/src/test/java/org/junit/jupiter/api/condition/EnabledIfEnvironmentVariableIntegrationTests.java`

`junit-jupiter-engine/src/test/java/org/junit/jupiter/api/condition/EnabledIfEnvironmentVariableIntegrationTests.java` lines 11-61
```
11: package org.junit.jupiter.api.condition;
12: 
13: import static org.junit.jupiter.api.Assertions.assertEquals;
14: import static org.junit.jupiter.api.Assertions.fail;
15: 
16: import java.lang.annotation.ElementType;
17: import java.lang.annotation.Retention;
18: import java.lang.annotation.RetentionPolicy;
19: import java.lang.annotation.Target;
20: 
21: import org.junit.jupiter.api.Disabled;
22: import org.junit.jupiter.api.Test;
23: 
24: /**
25:  * Integration tests for {@link EnabledIfEnvironmentVariable}.
26:  *
27:  * @since 5.1
28:  */
29: @Disabled("Disabled since the required environment variables are not set")
30: // Tests (except those with intentional configuration errors) will pass if you set
31: // the following environment variables:
32: // EnabledIfEnvironmentVariableTests.key1 = enigma
33: // EnabledIfEnvironmentVariableTests.key2 = enigma
34: class EnabledIfEnvironmentVariableIntegrationTests {
35: 
36: 	static final String KEY1 = "EnabledIfEnvironmentVariableTests.key1";
37: 	static final String KEY2 = "EnabledIfEnvironmentVariableTests.key2";
38: 	static final String ENIGMA = "enigma";
39: 	static final String PUZZLE = "puzzle";
40: 	static final String BOGUS = "bogus";
41: 
42: 	@Test
43: 	void enabledBecauseAnnotationIsNotPresent() {
44: 	}
45: 
46: 	@Test
47: 	@EnabledIfEnvironmentVariable(named = "  ", matches = ENIGMA)
48: 	void blankNamedAttribute() {
49: 	}
50: 
51: 	@Test
52: 	@EnabledIfEnvironmentVariable(named = KEY1, matches = "  ")
53: 	void blankMatchesAttribute() {
54: 	}
55: 
56: 	@Test
57: 	@EnabledIfEnvironmentVariable(named = KEY1, matches = ENIGMA)
58: 	void enabledBecauseEnvironmentVariableMatchesExactly() {
59: 		assertEquals(ENIGMA, System.getenv(KEY1));
60: 	}
61: 
```

- **A** (claude-opus-5-5): `false_positive` — rule: R3 Path context (also R1: mechanism absent) — The file is under `src/test/java/` and is clearly a test, and line 36 only holds an environment-variable name (`"EnabledIfEnvironmentVariableTests.key1"`), not a credential. The value "enigma" is a dummy test value, not an API key.
- **B** (gpt-6.1-sol): `false_positive` — rule: R3 Path context — The flagged file is under src/test/ and explicitly contains integration tests. The flagged string names a test environment variable, so the finding is in test code.

**Proposed label: `false_positive`**

Your decision: [ ] confirm  [ ] change to ______  note: ____

---

## 79c506c36b4a

- tool_signal: performance:ruff rule PERF401 at .github/actions/people/app/main.py:466
- path: `.github/actions/people/app/main.py`

`.github/actions/people/app/main.py` lines 441-491
```
441:     for issue in issue_nodes:
442:         issue_author_name = None
443:         if issue.author:
444:             authors[issue.author.login] = issue.author
445:             issue_author_name = issue.author.login
446:         issue_commentors = set()
447:         for comment in issue.comments.nodes:
448:             if comment.author:
449:                 authors[comment.author.login] = comment.author
450:                 if comment.author.login != issue_author_name:
451:                     issue_commentors.add(comment.author.login)
452:         for author_name in issue_commentors:
453:             commentors[author_name] += 1
454:             if issue.createdAt > one_month_ago:
455:                 last_month_commentors[author_name] += 1
456: 
457:     return commentors, last_month_commentors, authors
458: 
459: 
460: def get_discussions_experts(settings: Settings):
461:     discussion_nodes: List[DiscussionsNode] = []
462:     discussion_edges = get_graphql_question_discussion_edges(settings=settings)
463: 
464:     while discussion_edges:
465:         for discussion_edge in discussion_edges:
466:             discussion_nodes.append(discussion_edge.node)
467:         last_edge = discussion_edges[-1]
468:         discussion_edges = get_graphql_question_discussion_edges(
469:             settings=settings, after=last_edge.cursor
470:         )
471: 
472:     commentors = Counter()
473:     last_month_commentors = Counter()
474:     authors: Dict[str, Author] = {}
475: 
476:     now = datetime.now(tz=timezone.utc)
477:     one_month_ago = now - timedelta(days=30)
478: 
479:     for discussion in discussion_nodes:
480:         discussion_author_name = None
481:         if discussion.author:
482:             authors[discussion.author.login] = discussion.author
483:             discussion_author_name = discussion.author.login
484:         discussion_commentors = set()
485:         for comment in discussion.comments.nodes:
486:             if comment.author:
487:                 authors[comment.author.login] = comment.author
488:                 if comment.author.login != discussion_author_name:
489:                     discussion_commentors.add(comment.author.login)
490:             for reply in comment.replies.nodes:
491:                 if reply.author:
```

- **A** (claude-opus-5-5): `real_defect` — rule: R1 Presence — Line 466 literally shows the PERF401 pattern: a for-loop that only calls `discussion_nodes.append(discussion_edge.node)` instead of using `extend` or a comprehension. The path is hand-written CI action code, not a test, fixture, example, doc or generated file, and nothing in the excerpt guards or neutralises the pattern.
- **B** (gpt-6.1-sol): `real_defect` — rule: R1 Presence — Lines 465–466 visibly use a loop solely to append each discussion_edge.node to a list, the mechanism flagged by PERF401. No visible guard or neutralising construct applies, and the path does not indicate test or generated code.

**Proposed label: `real_defect`**

Your decision: [ ] confirm  [ ] change to ______  note: ____

---

## 7ac4363d7ef6

- tool_signal: security.semgrep.actions-mutable-ref rule scout.security.actions.mutable-ref at .github/workflows/build-docs.yml:16
- path: `.github/workflows/build-docs.yml`

`.github/workflows/build-docs.yml` lines 1-41
```
1: name: Build Docs
2: on:
3:   push:
4:     branches:
5:       - master
6:   pull_request:
7:     types: [opened, synchronize]
8: jobs:
9:   build-docs:
10:     runs-on: ubuntu-latest
11:     steps:
12:       - name: Dump GitHub context
13:         env:
14:           GITHUB_CONTEXT: ${{ toJson(github) }}
15:         run: echo "$GITHUB_CONTEXT"
16:       - uses: actions/checkout@v3
17:       - name: Set up Python
18:         uses: actions/setup-python@v4
19:         with:
20:           python-version: "3.11"
21:       - uses: actions/cache@v3
22:         id: cache
23:         with:
24:           path: ${{ env.pythonLocation }}
25:           key: ${{ runner.os }}-python-docs-${{ env.pythonLocation }}-${{ hashFiles('pyproject.toml') }}-v03
26:       - name: Install docs extras
27:         if: steps.cache.outputs.cache-hit != 'true'
28:         run: pip install .[doc]
29:       - name: Install Material for MkDocs Insiders
30:         if: ( github.event_name != 'pull_request' || github.event.pull_request.head.repo.fork == false ) && steps.cache.outputs.cache-hit != 'true'
31:         run: pip install git+https://${{ secrets.ACTIONS_TOKEN }}@github.com/squidfunk/mkdocs-material-insiders.git
32:       - name: Build Docs
33:         run: python ./scripts/docs.py build-all
34:       - name: Zip docs
35:         run: bash ./scripts/zip-docs.sh
36:       - uses: actions/upload-artifact@v3
37:         with:
38:           name: docs-zip
39:           path: ./site/docs.zip
40:       - name: Deploy to Netlify
41:         uses: nwtgck/actions-netlify@v2.0.0
```

- **A** (claude-opus-5-5): `real_defect` — rule: R1 Presence (R3 does not apply) — Line 16 pins `actions/checkout@v3`, which is a mutable tag rather than a commit SHA, and nothing in the excerpt guards or neutralises it. The path is a real CI workflow, not a docs file, because only the file name contains the word "docs".
- **B** (gpt-6.1-sol): `real_defect` — rule: R1 Presence — Line 16 uses actions/checkout@v3, a mutable tag rather than an immutable commit pin. No visible guard or neutralising construct addresses the mutable reference.

**Proposed label: `real_defect`**

Your decision: [ ] confirm  [ ] change to ______  note: ____

---

## 802e9168665f

- tool_signal: security.semgrep.actions-mutable-ref rule scout.security.actions.mutable-ref at .github/workflows/richnav.yml:20
- path: `.github/workflows/richnav.yml`

`.github/workflows/richnav.yml` lines 1-25
```
1: name: eShopOnWeb - Code Index
2: 
3: on: workflow_dispatch
4: 
5: jobs:
6:   build:
7: 
8:     runs-on: windows-latest
9: 
10:     steps:
11:     - uses: actions/checkout@v2
12:     - name: Setup .NET Core
13:       uses: actions/setup-dotnet@v1
14:       with:
15:         dotnet-version: 8.0.x
16: 
17:     - name: Build with dotnet
18:       run: dotnet build ./Everything.sln --configuration Release /bl
19: 
20:     - uses: microsoft/RichCodeNavIndexer@v0.1
21:       with:
22:         repo-token: ${{ github.token }}
23:         languages: 'csharp'
24:         environment: 'internal'
25: 
```

- **A** (claude-opus-5-5): `real_defect` — rule: R1 Presence (no R2/R3 guard applies) — Line 20 uses `microsoft/RichCodeNavIndexer@v0.1`, a mutable tag rather than a pinned commit SHA, in a shipped workflow file. It also receives `github.token`, and nothing in the excerpt pins or otherwise neutralises the mutable ref.
- **B** (gpt-6.1-sol): `real_defect` — rule: R1 Presence — Line 20 references microsoft/RichCodeNavIndexer using the mutable tag @v0.1 rather than an immutable commit SHA. No visible pin or guard neutralises the mechanism.

**Proposed label: `real_defect`**

Your decision: [ ] confirm  [ ] change to ______  note: ____

---

## 94822c8d4cbc

- tool_signal: security.gitleaks.default-credentials rule jwt at tests/test_tutorial/test_security/test_tutorial005_an.py:306
- path: `tests/test_tutorial/test_security/test_tutorial005_an.py`

`tests/test_tutorial/test_security/test_tutorial005_an.py` lines 281-331
```
281:     response = client.get(
282:         "/users/me",
283:         headers={
284:             "Authorization": "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJmb28ifQ.NnExK_dlNAYyzACrXtXDrcWOgGY2JuPbI4eDaHdfK5Y"
285:         },
286:     )
287:     assert response.status_code == 401, response.text
288:     assert response.json() == {"detail": "Could not validate credentials"}
289:     assert response.headers["WWW-Authenticate"] == 'Bearer scope="me"'
290: 
291: 
292: def test_token_no_scope():
293:     access_token = get_access_token()
294:     response = client.get(
295:         "/users/me", headers={"Authorization": f"Bearer {access_token}"}
296:     )
297:     assert response.status_code == 401, response.text
298:     assert response.json() == {"detail": "Not enough permissions"}
299:     assert response.headers["WWW-Authenticate"] == 'Bearer scope="me"'
300: 
301: 
302: def test_token_inexistent_user():
303:     response = client.get(
304:         "/users/me",
305:         headers={
306:             "Authorization": "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VybmFtZTpib2IifQ.HcfCW67Uda-0gz54ZWTqmtgJnZeNem0Q757eTa9EZuw"
307:         },
308:     )
309:     assert response.status_code == 401, response.text
310:     assert response.json() == {"detail": "Could not validate credentials"}
311:     assert response.headers["WWW-Authenticate"] == 'Bearer scope="me"'
312: 
313: 
314: def test_token_inactive_user():
315:     access_token = get_access_token(
316:         username="alice", password="secretalice", scope="me"
317:     )
318:     response = client.get(
319:         "/users/me", headers={"Authorization": f"Bearer {access_token}"}
320:     )
321:     assert response.status_code == 400, response.text
322:     assert response.json() == {"detail": "Inactive user"}
323: 
324: 
325: def test_read_items():
326:     access_token = get_access_token(scope="me items")
327:     response = client.get(
328:         "/users/me/items/", headers={"Authorization": f"Bearer {access_token}"}
329:     )
330:     assert response.status_code == 200, response.text
331:     assert response.json() == [{"item_id": "Foo", "owner": "johndoe"}]
```

- **A** (claude-opus-5-5): `false_positive` — rule: R3 Path context — The flagged JWT at line 306 is literally present, but only in a test file under `tests/test_tutorial/test_security/`, where it is a deliberately invalid token for a nonexistent user that the test expects to be rejected with 401. The mechanism appears only in test code, so R3 makes it a false positive.
- **B** (gpt-6.1-sol): `false_positive` — rule: R3 Path context — The JWT appears in a test file under tests/ and is used in test_token_inexistent_user to verify credential rejection.

**Proposed label: `false_positive`**

Your decision: [ ] confirm  [ ] change to ______  note: ____

---

## 94bdc758adf8

- tool_signal: performance:ruff rule PERF401 at fastapi/encoders.py:130
- path: `fastapi/encoders.py`

`fastapi/encoders.py` lines 105-155
```
105:                 )
106:                 and (value is not None or not exclude_none)
107:                 and key in allowed_keys
108:             ):
109:                 encoded_key = jsonable_encoder(
110:                     key,
111:                     by_alias=by_alias,
112:                     exclude_unset=exclude_unset,
113:                     exclude_none=exclude_none,
114:                     custom_encoder=custom_encoder,
115:                     sqlalchemy_safe=sqlalchemy_safe,
116:                 )
117:                 encoded_value = jsonable_encoder(
118:                     value,
119:                     by_alias=by_alias,
120:                     exclude_unset=exclude_unset,
121:                     exclude_none=exclude_none,
122:                     custom_encoder=custom_encoder,
123:                     sqlalchemy_safe=sqlalchemy_safe,
124:                 )
125:                 encoded_dict[encoded_key] = encoded_value
126:         return encoded_dict
127:     if isinstance(obj, (list, set, frozenset, GeneratorType, tuple)):
128:         encoded_list = []
129:         for item in obj:
130:             encoded_list.append(
131:                 jsonable_encoder(
132:                     item,
133:                     include=include,
134:                     exclude=exclude,
135:                     by_alias=by_alias,
136:                     exclude_unset=exclude_unset,
137:                     exclude_defaults=exclude_defaults,
138:                     exclude_none=exclude_none,
139:                     custom_encoder=custom_encoder,
140:                     sqlalchemy_safe=sqlalchemy_safe,
141:                 )
142:             )
143:         return encoded_list
144: 
145:     if type(obj) in ENCODERS_BY_TYPE:
146:         return ENCODERS_BY_TYPE[type(obj)](obj)
147:     for encoder, classes_tuple in encoders_by_class_tuples.items():
148:         if isinstance(obj, classes_tuple):
149:             return encoder(obj)
150: 
151:     try:
152:         data = dict(obj)
153:     except Exception as e:
154:         errors: List[Exception] = []
155:         errors.append(e)
```

- **A** (claude-opus-5-5): `real_defect` — rule: R1 Presence — Lines 128-143 literally show the PERF401 pattern: a for-loop that appends to `encoded_list` and could be a list comprehension. It is in shipped library code (fastapi/encoders.py), and nothing in the excerpt guards or neutralises it.
- **B** (gpt-6.1-sol): `real_defect` — rule: R1 Presence — Lines 128–142 visibly build a list with a for loop and append, the mechanism flagged by PERF401. This is hand-written shipped code, with no visible construct neutralising the finding.

**Proposed label: `real_defect`**

Your decision: [ ] confirm  [ ] change to ______  note: ____

---

## 9ba6cc792b32

- tool_signal: security.gitleaks.default-credentials rule generic-api-key at junit-jupiter-engine/src/test/java/org/junit/jupiter/api/condition/EnabledIfEnvironmentVariableIntegrationTests.java:37
- path: `junit-jupiter-engine/src/test/java/org/junit/jupiter/api/condition/EnabledIfEnvironmentVariableIntegrationTests.java`

`junit-jupiter-engine/src/test/java/org/junit/jupiter/api/condition/EnabledIfEnvironmentVariableIntegrationTests.java` lines 12-62
```
12: 
13: import static org.junit.jupiter.api.Assertions.assertEquals;
14: import static org.junit.jupiter.api.Assertions.fail;
15: 
16: import java.lang.annotation.ElementType;
17: import java.lang.annotation.Retention;
18: import java.lang.annotation.RetentionPolicy;
19: import java.lang.annotation.Target;
20: 
21: import org.junit.jupiter.api.Disabled;
22: import org.junit.jupiter.api.Test;
23: 
24: /**
25:  * Integration tests for {@link EnabledIfEnvironmentVariable}.
26:  *
27:  * @since 5.1
28:  */
29: @Disabled("Disabled since the required environment variables are not set")
30: // Tests (except those with intentional configuration errors) will pass if you set
31: // the following environment variables:
32: // EnabledIfEnvironmentVariableTests.key1 = enigma
33: // EnabledIfEnvironmentVariableTests.key2 = enigma
34: class EnabledIfEnvironmentVariableIntegrationTests {
35: 
36: 	static final String KEY1 = "EnabledIfEnvironmentVariableTests.key1";
37: 	static final String KEY2 = "EnabledIfEnvironmentVariableTests.key2";
38: 	static final String ENIGMA = "enigma";
39: 	static final String PUZZLE = "puzzle";
40: 	static final String BOGUS = "bogus";
41: 
42: 	@Test
43: 	void enabledBecauseAnnotationIsNotPresent() {
44: 	}
45: 
46: 	@Test
47: 	@EnabledIfEnvironmentVariable(named = "  ", matches = ENIGMA)
48: 	void blankNamedAttribute() {
49: 	}
50: 
51: 	@Test
52: 	@EnabledIfEnvironmentVariable(named = KEY1, matches = "  ")
53: 	void blankMatchesAttribute() {
54: 	}
55: 
56: 	@Test
57: 	@EnabledIfEnvironmentVariable(named = KEY1, matches = ENIGMA)
58: 	void enabledBecauseEnvironmentVariableMatchesExactly() {
59: 		assertEquals(ENIGMA, System.getenv(KEY1));
60: 	}
61: 
62: 	@Test
```

- **A** (claude-opus-5-5): `false_positive` — rule: R3 Path context (also R1: mechanism absent) — The file is under `src/test/java/` and the class is an integration test. Line 37 is just an environment-variable name string (`"EnabledIfEnvironmentVariableTests.key2"`), not a credential or API key.
- **B** (gpt-6.1-sol): `false_positive` — rule: R3 Path context — The file is under src/test/ and explicitly contains integration tests. The flagged constant names a test environment variable.

**Proposed label: `false_positive`**

Your decision: [ ] confirm  [ ] change to ______  note: ____

---

## architecture-routing-utils-cycle-d666ccb6

- tool_signal: architecture analyzer rule ARCHITECTURE007: circular import between fastapi/routing.py and fastapi/utils.py
- path: `fastapi/routing.py`

`fastapi/routing.py` lines 20-56
```
20: )
21: 
22: from fastapi import params
23: from fastapi.datastructures import Default, DefaultPlaceholder
24: from fastapi.dependencies.models import Dependant
25: from fastapi.dependencies.utils import (
26:     get_body_field,
27:     get_dependant,
28:     get_parameterless_sub_dependant,
29:     get_typed_return_annotation,
30:     solve_dependencies,
31: )
32: from fastapi.encoders import DictIntStrAny, SetIntStr, jsonable_encoder
33: from fastapi.exceptions import RequestValidationError, WebSocketRequestValidationError
34: from fastapi.types import DecoratedCallable
35: from fastapi.utils import (
36:     create_cloned_field,
37:     create_response_field,
38:     generate_unique_id,
39:     get_value_or_default,
40:     is_body_allowed_for_status_code,
41: )
42: from pydantic import BaseModel
43: from pydantic.error_wrappers import ErrorWrapper, ValidationError
44: from pydantic.fields import ModelField, Undefined
45: from pydantic.utils import lenient_issubclass
46: from starlette import routing
47: from starlette.concurrency import run_in_threadpool
48: from starlette.exceptions import HTTPException
49: from starlette.requests import Request
50: from starlette.responses import JSONResponse, Response
51: from starlette.routing import BaseRoute, Match
52: from starlette.routing import Mount as Mount  # noqa
53: from starlette.routing import (
54:     compile_path,
55:     get_name,
56:     request_response,
```

`fastapi/utils.py` lines 1-32
```
1: import re
2: import warnings
3: from dataclasses import is_dataclass
4: from enum import Enum
5: from typing import TYPE_CHECKING, Any, Dict, Optional, Set, Type, Union, cast
6: 
7: import fastapi
8: from fastapi.datastructures import DefaultPlaceholder, DefaultType
9: from fastapi.openapi.constants import REF_PREFIX
10: from pydantic import BaseConfig, BaseModel, create_model
11: from pydantic.class_validators import Validator
12: from pydantic.fields import FieldInfo, ModelField, UndefinedType
13: from pydantic.schema import model_process_schema
14: from pydantic.utils import lenient_issubclass
15: 
16: if TYPE_CHECKING:  # pragma: nocover
17:     from .routing import APIRoute
18: 
19: 
20: def is_body_allowed_for_status_code(status_code: Union[int, str, None]) -> bool:
21:     if status_code is None:
22:         return True
23:     # Ref: https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.1.0.md#patterned-fields-1
24:     if status_code in {
25:         "default",
26:         "1XX",
27:         "2XX",
28:         "3XX",
29:         "4XX",
30:         "5XX",
31:     }:
32:         return True
```

- **A** (claude-opus-5-5): `false_positive` — rule: R4 Type-only — routing.py imports fastapi.utils at runtime, but utils.py imports `.routing` only inside an `if TYPE_CHECKING:` block, so the back-edge is type-only. That means there is no runtime circular import between the two modules.
- **B** (gpt-6.1-sol): `false_positive` — rule: R4 Type-only — The reverse import from fastapi/utils.py to routing.APIRoute is inside an if TYPE_CHECKING block, so it does not create a runtime circular import.

**Proposed label: `false_positive`**

Your decision: [ ] confirm  [ ] change to ______  note: ____

---

## ca27b4817509

- tool_signal: security.gitleaks.default-credentials rule jwt at examples/jwt/src/main.rs:40
- path: `examples/jwt/src/main.rs`

`examples/jwt/src/main.rs` lines 15-65
```
15:     routing::{get, post},
16:     Json, RequestPartsExt, Router,
17: };
18: use jsonwebtoken::{decode, encode, DecodingKey, EncodingKey, Header, Validation};
19: use once_cell::sync::Lazy;
20: use serde::{Deserialize, Serialize};
21: use serde_json::json;
22: use std::{fmt::Display, net::SocketAddr};
23: use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};
24: 
25: // Quick instructions
26: //
27: // - get an authorization token:
28: //
29: // curl -s \
30: //     -w '\n' \
31: //     -H 'Content-Type: application/json' \
32: //     -d '{"client_id":"foo","client_secret":"bar"}' \
33: //     http://localhost:3000/authorize
34: //
35: // - visit the protected area using the authorized token
36: //
37: // curl -s \
38: //     -w '\n' \
39: //     -H 'Content-Type: application/json' \
40: //     -H 'Authorization: Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJiQGIuY29tIiwiY29tcGFueSI6IkFDTUUiLCJleHAiOjEwMDAwMDAwMDAwfQ.M3LAZmrzUkXDC1q5mSzFAs_kJrwuKz3jOoDmjJ0G4gM' \
41: //     http://localhost:3000/protected
42: //
43: // - try to visit the protected area using an invalid token
44: //
45: // curl -s \
46: //     -w '\n' \
47: //     -H 'Content-Type: application/json' \
48: //     -H 'Authorization: Bearer blahblahblah' \
49: //     http://localhost:3000/protected
50: 
51: static KEYS: Lazy<Keys> = Lazy::new(|| {
52:     let secret = std::env::var("JWT_SECRET").expect("JWT_SECRET must be set");
53:     Keys::new(secret.as_bytes())
54: });
55: 
56: #[tokio::main]
57: async fn main() {
58:     tracing_subscriber::registry()
59:         .with(
60:             tracing_subscriber::EnvFilter::try_from_default_env()
61:                 .unwrap_or_else(|_| "example_jwt=debug".into()),
62:         )
63:         .with(tracing_subscriber::fmt::layer())
64:         .init();
65: 
```

- **A** (claude-opus-5-5): `false_positive` — rule: R3 Path context (and R2: the actual signing secret comes from the JWT_SECRET environment variable at line 52) — The flagged line 40 is a JWT inside a usage-instruction comment in the examples/ directory, which is example code and not shipped code. The real secret is read from the JWT_SECRET environment variable, not hardcoded.
- **B** (gpt-6.1-sol): `false_positive` — rule: R3 Path context — The JWT appears in a commented usage example in examples/jwt/src/main.rs. The finding is confined to an example file, so R3 applies.

**Proposed label: `false_positive`**

Your decision: [ ] confirm  [ ] change to ______  note: ____

---

## csharp-architecture-generated-migration-length-4da8212

- tool_signal: architecture patterns analyzer rule ARCHITECTURE060: method too long (src/Infrastructure/Data/Migrations/20201202111507_InitialModel.cs)
- path: `src/Infrastructure/Data/Migrations/20201202111507_InitialModel.cs`

`src/Infrastructure/Data/Migrations/20201202111507_InitialModel.cs` lines 1-45
```
1: ﻿using System;
2: using Microsoft.EntityFrameworkCore.Migrations;
3: 
4: namespace Microsoft.eShopWeb.Infrastructure.Data.Migrations;
5: 
6: public partial class InitialModel : Migration
7: {
8:     protected override void Up(MigrationBuilder migrationBuilder)
9:     {
10:         migrationBuilder.CreateSequence(
11:             name: "catalog_brand_hilo",
12:             incrementBy: 10);
13: 
14:         migrationBuilder.CreateSequence(
15:             name: "catalog_hilo",
16:             incrementBy: 10);
17: 
18:         migrationBuilder.CreateSequence(
19:             name: "catalog_type_hilo",
20:             incrementBy: 10);
21: 
22:         migrationBuilder.CreateTable(
23:             name: "Baskets",
24:             columns: table => new
25:             {
26:                 Id = table.Column<int>(type: "int", nullable: false)
27:                     .Annotation("SqlServer:Identity", "1, 1"),
28:                 BuyerId = table.Column<string>(type: "nvarchar(40)", maxLength: 40, nullable: false)
29:             },
30:             constraints: table =>
31:             {
32:                 table.PrimaryKey("PK_Baskets", x => x.Id);
33:             });
34: 
35:         migrationBuilder.CreateTable(
36:             name: "CatalogBrands",
37:             columns: table => new
38:             {
39:                 Id = table.Column<int>(type: "int", nullable: false),
40:                 Brand = table.Column<string>(type: "nvarchar(100)", maxLength: 100, nullable: false)
41:             },
42:             constraints: table =>
43:             {
44:                 table.PrimaryKey("PK_CatalogBrands", x => x.Id);
45:             });
```

- **A** (claude-opus-5-5): `false_positive` — rule: R5 Metrics (with R3 generated-file context) — The excerpt shows the method length comes from repetitive declarative EF Core migration builder calls (CreateSequence/CreateTable) in a timestamped, tool-generated migration file. Its length does not reflect real maintainability cost.
- **B** (gpt-6.1-sol): `false_positive` — rule: R5 Metrics — The excerpt shows repetitive declarative migration operations defining sequences, tables, columns, and constraints. This inflates method length without showing complex hand-written logic.

**Proposed label: `false_positive`**

Your decision: [ ] confirm  [ ] change to ______  note: ____

---

## csharp-performance-identity-height-transition-4da8212

- tool_signal: frontend performance analyzer rules PERFORMANCE013-PERFORMANCE016: CSS transition on layout-affecting property (height)
- path: `src/Web/wwwroot/css/shared/components/identity/identity.css`

`src/Web/wwwroot/css/shared/components/identity/identity.css` lines 15-59
```
15: 
16: .esh-identity-name--upper {
17:   text-transform: uppercase;
18: }
19: 
20: @media screen and (max-width: 768px) {
21:   .esh-identity-name {
22:     font-size: 0.85rem;
23:   }
24: }
25: 
26: .esh-identity-image {
27:   display: inline-block;
28: }
29: 
30: .esh-identity-drop {
31:   background: #FFFFFF;
32:   height: 10px;
33:   width: 10rem;
34:   overflow: hidden;
35:   padding: .5rem;
36:   position: absolute;
37:   right: 0;
38:   top: 2.5rem;
39:   transition: height 0.35s;
40: }
41: 
42: .esh-identity:hover .esh-identity-drop {
43:   border: 1px solid #EEEEEE;
44:   height: 14rem;
45:   transition: height 0.35s;
46:   z-index: 10;
47: }
48: 
49: .esh-identity-item {
50:   cursor: pointer;
51:   transition: color 0.35s;
52: }
53: 
54: .esh-identity-item:hover {
55:   color: #75b918;
56:   transition: color 0.35s;
57: }
58: 
59: 
```

`src/Web/Views/Shared/_Layout.cshtml` lines 10-35
```
10:         <link rel="stylesheet" href="~/css/app.component.css" />
11:         <link rel="stylesheet" href="~/css/shared/components/header/header.css" />
12:         <link rel="stylesheet" href="~/css/shared/components/identity/identity.css" />
13:         <link rel="stylesheet" href="~/css/shared/components/pager/pager.css" />
14:         <link rel="stylesheet" href="~/css/basket/basket.component.css" />
15:         <link rel="stylesheet" href="~/css/basket/basket-status/basket-status.component.css" />
16:         <link rel="stylesheet" href="~/css/catalog/catalog.component.css" />
17:         <link rel="stylesheet" href="~/css/orders/orders.component.css" />
18:     </environment>
19:     <environment names="Staging,Production">
20:         <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/twitter-bootstrap/4.0.0-alpha.5/css/bootstrap.min.css"
21:               asp-fallback-href="~/lib/bootstrap/dist/css/bootstrap.min.css"
22:               asp-fallback-test-class="sr-only" asp-fallback-test-property="position" asp-fallback-test-value="absolute" />
23:         <link rel="stylesheet" href="~/css/site.min.css" asp-append-version="true" />
24:     </environment>
25: </head>
26: <body>
27:     <div class="esh-app-wrapper">
28:         <header class="esh-app-header">
29:             <div class="container">
30:                 <article class="row">
31:                     <section class="col-lg-7 col-md-6 col-xs-12">
32:                         <a asp-area="" asp-page="/Index">
33:                             <img src="~/images/brand.png" alt="eShop On Web" />
34:                         </a>
35:                     </section>
```

- **A** (claude-opus-5-5): `real_defect` — rule: R1 Presence (no R2–R5 false_positive rule applies) — Lines 39 and 45 have the literal declaration `transition: height 0.35s` on `.esh-identity-drop`, which is shipped hand-written CSS loaded by _Layout.cshtml, so the flagged mechanism is present. Nothing in the excerpt guards or neutralises it, and the path is not test, fixture, generated or vendor code.
- **B** (gpt-6.1-sol): `real_defect` — rule: R1 Presence — The CSS explicitly transitions height from 10px to 14rem on hover in shipped, non-test code. No visible guard or neutralising construct addresses the layout-affecting transition.

**Proposed label: `real_defect`**

Your decision: [ ] confirm  [ ] change to ______  note: ____

---

## csharp-quality-generated-migration-length-4da8212

- tool_signal: lizard rule COMPLEXITY008: function length 152 lines, cyclomatic complexity 1 (high severity)
- path: `src/Infrastructure/Data/Migrations/20201202111507_InitialModel.cs`

`src/Infrastructure/Data/Migrations/20201202111507_InitialModel.cs` lines 1-45
```
1: ﻿using System;
2: using Microsoft.EntityFrameworkCore.Migrations;
3: 
4: namespace Microsoft.eShopWeb.Infrastructure.Data.Migrations;
5: 
6: public partial class InitialModel : Migration
7: {
8:     protected override void Up(MigrationBuilder migrationBuilder)
9:     {
10:         migrationBuilder.CreateSequence(
11:             name: "catalog_brand_hilo",
12:             incrementBy: 10);
13: 
14:         migrationBuilder.CreateSequence(
15:             name: "catalog_hilo",
16:             incrementBy: 10);
17: 
18:         migrationBuilder.CreateSequence(
19:             name: "catalog_type_hilo",
20:             incrementBy: 10);
21: 
22:         migrationBuilder.CreateTable(
23:             name: "Baskets",
24:             columns: table => new
25:             {
26:                 Id = table.Column<int>(type: "int", nullable: false)
27:                     .Annotation("SqlServer:Identity", "1, 1"),
28:                 BuyerId = table.Column<string>(type: "nvarchar(40)", maxLength: 40, nullable: false)
29:             },
30:             constraints: table =>
31:             {
32:                 table.PrimaryKey("PK_Baskets", x => x.Id);
33:             });
34: 
35:         migrationBuilder.CreateTable(
36:             name: "CatalogBrands",
37:             columns: table => new
38:             {
39:                 Id = table.Column<int>(type: "int", nullable: false),
40:                 Brand = table.Column<string>(type: "nvarchar(100)", maxLength: 100, nullable: false)
41:             },
42:             constraints: table =>
43:             {
44:                 table.PrimaryKey("PK_CatalogBrands", x => x.Id);
45:             });
```

- **A** (claude-opus-5-5): `false_positive` — rule: R5 Metrics (also R3: EF Core migration is tool-generated) — The excerpt shows the function length comes from repetitive declarative CreateSequence/CreateTable migration builder calls with cyclomatic complexity 1, so the metric does not reflect real maintainability cost. The file is an auto-generated EF Core migration (timestamped Migrations/ file), not hand-written logic.
- **B** (gpt-6.1-sol): `false_positive` — rule: R5 Metrics — The excerpt shows repetitive declarative migration operations defining sequences, tables, columns, and constraints. This inflates function length without showing complex hand-written logic, consistent with the reported cyclomatic complexity of 1.

**Proposed label: `false_positive`**

Your decision: [ ] confirm  [ ] change to ______  note: ____

---

## csharp-security-mutable-actions-4da8212

- tool_signal: semgrep rules SECURITY002, SECURITY003: GitHub Action referenced by version tag instead of commit SHA
- path: `.github/workflows/dotnetcore.yml`

`.github/workflows/dotnetcore.yml` lines 1-23
```
1: name: eShopOnWeb Build and Test
2: 
3: on: [push, pull_request, workflow_dispatch]
4: 
5: jobs:
6:   build:
7: 
8:     runs-on: ubuntu-latest
9: 
10:     steps:
11:     - uses: actions/checkout@v2
12:     - name: Setup .NET
13:       uses: actions/setup-dotnet@v1
14:       with:
15:         dotnet-version: '8.0.x'
16:         include-prerelease: true
17: 
18:     - name: Build with dotnet
19:       run: dotnet build ./eShopOnWeb.sln --configuration Release
20:     
21:     - name: Test with dotnet
22:       run: dotnet test ./eShopOnWeb.sln --configuration Release
23: 
```

- **A** (claude-opus-5-5): `real_defect` — rule: R1 Presence (no false_positive rule R2–R5 applies) — Lines 11 and 13 literally reference `actions/checkout@v2` and `actions/setup-dotnet@v1` by version tag rather than commit SHA in a shipped CI workflow. No pin, guard or test/fixture path context is visible in the excerpt.
- **B** (gpt-6.1-sol): `real_defect` — rule: R1 Presence — The workflow references actions/checkout@v2 and actions/setup-dotnet@v1 using version tags instead of commit SHAs. No visible pin or neutralising guard addresses this mechanism.

**Proposed label: `real_defect`**

Your decision: [ ] confirm  [ ] change to ______  note: ____

---

## d6269e9ba7ee

- tool_signal: security.semgrep.actions-mutable-ref rule scout.security.actions.mutable-ref at .github/workflows/test.yml:26
- path: `.github/workflows/test.yml`

`.github/workflows/test.yml` lines 1-51
```
1: name: Test
2: 
3: on:
4:   push:
5:     branches:
6:       - master
7:   pull_request:
8:     types: [opened, synchronize]
9: 
10: jobs:
11:   test:
12:     runs-on: ubuntu-latest
13:     strategy:
14:       matrix:
15:         python-version: ["3.7", "3.8", "3.9", "3.10", "3.11"]
16:       fail-fast: false
17: 
18:     steps:
19:       - uses: actions/checkout@v3
20:       - name: Set up Python
21:         uses: actions/setup-python@v4
22:         with:
23:           python-version: ${{ matrix.python-version }}
24:           cache: "pip"
25:           cache-dependency-path: pyproject.toml
26:       - uses: actions/cache@v3
27:         id: cache
28:         with:
29:           path: ${{ env.pythonLocation }}
30:           key: ${{ runner.os }}-python-${{ env.pythonLocation }}-${{ hashFiles('pyproject.toml') }}-test-v03
31:       - name: Install Dependencies
32:         if: steps.cache.outputs.cache-hit != 'true'
33:         run: pip install -e .[all,dev,doc,test]
34:       - name: Lint
35:         run: bash scripts/lint.sh
36:       - run: mkdir coverage
37:       - name: Test
38:         run: bash scripts/test.sh
39:         env:
40:           COVERAGE_FILE: coverage/.coverage.${{ runner.os }}-py${{ matrix.python-version }}
41:           CONTEXT: ${{ runner.os }}-py${{ matrix.python-version }}
42:       - name: Store coverage files
43:         uses: actions/upload-artifact@v3
44:         with:
45:           name: coverage
46:           path: coverage
47:   coverage-combine:
48:     needs: [test]
49:     runs-on: ubuntu-latest
50: 
51:     steps:
```

- **A** (claude-opus-5-5): `real_defect` — rule: R1 Presence — Line 26 literally uses `actions/cache@v3`, a mutable tag rather than a pinned commit SHA, in a shipped CI workflow (not a test/fixture path). Nothing in the excerpt pins or otherwise neutralises the mutable reference.
- **B** (gpt-6.1-sol): `real_defect` — rule: R1 Presence — Line 26 uses actions/cache@v3, a mutable tag rather than an immutable commit pin. The workflow is shipped automation, and no visible guard or pin neutralises the mechanism.

**Proposed label: `real_defect`**

Your decision: [ ] confirm  [ ] change to ______  note: ____

---

## d648aace50e1

- tool_signal: security.gitleaks.default-credentials rule curl-auth-header at examples/jwt/src/main.rs:37
- path: `examples/jwt/src/main.rs`

`examples/jwt/src/main.rs` lines 12-62
```
12:     headers::{authorization::Bearer, Authorization},
13:     http::{request::Parts, StatusCode},
14:     response::{IntoResponse, Response},
15:     routing::{get, post},
16:     Json, RequestPartsExt, Router,
17: };
18: use jsonwebtoken::{decode, encode, DecodingKey, EncodingKey, Header, Validation};
19: use once_cell::sync::Lazy;
20: use serde::{Deserialize, Serialize};
21: use serde_json::json;
22: use std::{fmt::Display, net::SocketAddr};
23: use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};
24: 
25: // Quick instructions
26: //
27: // - get an authorization token:
28: //
29: // curl -s \
30: //     -w '\n' \
31: //     -H 'Content-Type: application/json' \
32: //     -d '{"client_id":"foo","client_secret":"bar"}' \
33: //     http://localhost:3000/authorize
34: //
35: // - visit the protected area using the authorized token
36: //
37: // curl -s \
38: //     -w '\n' \
39: //     -H 'Content-Type: application/json' \
40: //     -H 'Authorization: Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJiQGIuY29tIiwiY29tcGFueSI6IkFDTUUiLCJleHAiOjEwMDAwMDAwMDAwfQ.M3LAZmrzUkXDC1q5mSzFAs_kJrwuKz3jOoDmjJ0G4gM' \
41: //     http://localhost:3000/protected
42: //
43: // - try to visit the protected area using an invalid token
44: //
45: // curl -s \
46: //     -w '\n' \
47: //     -H 'Content-Type: application/json' \
48: //     -H 'Authorization: Bearer blahblahblah' \
49: //     http://localhost:3000/protected
50: 
51: static KEYS: Lazy<Keys> = Lazy::new(|| {
52:     let secret = std::env::var("JWT_SECRET").expect("JWT_SECRET must be set");
53:     Keys::new(secret.as_bytes())
54: });
55: 
56: #[tokio::main]
57: async fn main() {
58:     tracing_subscriber::registry()
59:         .with(
60:             tracing_subscriber::EnvFilter::try_from_default_env()
61:                 .unwrap_or_else(|_| "example_jwt=debug".into()),
62:         )
```

- **A** (claude-opus-5-5): `false_positive` — rule: R3 Path context — The flagged curl Authorization header at line 37 is only in a code comment showing how to use the example, and the file is under the `examples/` directory. The real secret is read from the `JWT_SECRET` environment variable, not hardcoded.
- **B** (gpt-6.1-sol): `false_positive` — rule: R3 Path context — The flagged curl authorization header appears in commented instructions in examples/jwt/src/main.rs, a clearly identified example file.

**Proposed label: `false_positive`**

Your decision: [ ] confirm  [ ] change to ______  note: ____

---

## e2c409a78bc0

- tool_signal: security.gitleaks.default-credentials rule jwt at tests/test_tutorial/test_security/test_tutorial005_py39.py:330
- path: `tests/test_tutorial/test_security/test_tutorial005_py39.py`

`tests/test_tutorial/test_security/test_tutorial005_py39.py` lines 305-355
```
305:         headers={
306:             "Authorization": "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJmb28ifQ.NnExK_dlNAYyzACrXtXDrcWOgGY2JuPbI4eDaHdfK5Y"
307:         },
308:     )
309:     assert response.status_code == 401, response.text
310:     assert response.json() == {"detail": "Could not validate credentials"}
311:     assert response.headers["WWW-Authenticate"] == 'Bearer scope="me"'
312: 
313: 
314: @needs_py39
315: def test_token_no_scope(client: TestClient):
316:     access_token = get_access_token(client=client)
317:     response = client.get(
318:         "/users/me", headers={"Authorization": f"Bearer {access_token}"}
319:     )
320:     assert response.status_code == 401, response.text
321:     assert response.json() == {"detail": "Not enough permissions"}
322:     assert response.headers["WWW-Authenticate"] == 'Bearer scope="me"'
323: 
324: 
325: @needs_py39
326: def test_token_inexistent_user(client: TestClient):
327:     response = client.get(
328:         "/users/me",
329:         headers={
330:             "Authorization": "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VybmFtZTpib2IifQ.HcfCW67Uda-0gz54ZWTqmtgJnZeNem0Q757eTa9EZuw"
331:         },
332:     )
333:     assert response.status_code == 401, response.text
334:     assert response.json() == {"detail": "Could not validate credentials"}
335:     assert response.headers["WWW-Authenticate"] == 'Bearer scope="me"'
336: 
337: 
338: @needs_py39
339: def test_token_inactive_user(client: TestClient):
340:     access_token = get_access_token(
341:         username="alice", password="secretalice", scope="me", client=client
342:     )
343:     response = client.get(
344:         "/users/me", headers={"Authorization": f"Bearer {access_token}"}
345:     )
346:     assert response.status_code == 400, response.text
347:     assert response.json() == {"detail": "Inactive user"}
348: 
349: 
350: @needs_py39
351: def test_read_items(client: TestClient):
352:     access_token = get_access_token(scope="me items", client=client)
353:     response = client.get(
354:         "/users/me/items/", headers={"Authorization": f"Bearer {access_token}"}
355:     )
```

- **A** (claude-opus-5-5): `false_positive` — rule: R3 Path context — The JWT at line 330 is literally present, but only in a test file under `tests/test_tutorial/`, where it serves as a token that the test expects to be rejected (it asserts a 401 'Could not validate credentials'). The mechanism appears only in test code, so R3 makes this a false_positive.
- **B** (gpt-6.1-sol): `false_positive` — rule: R3 Path context — The JWT appears in a test file under tests/ and is used in a test asserting rejection of credentials for a nonexistent user.

**Proposed label: `false_positive`**

Your decision: [ ] confirm  [ ] change to ______  note: ____

---

## ea0aeeb89124

- tool_signal: security.gitleaks.default-credentials rule jwt at tests/test_tutorial/test_security/test_tutorial005_an_py39.py:293
- path: `tests/test_tutorial/test_security/test_tutorial005_an_py39.py`

`tests/test_tutorial/test_security/test_tutorial005_an_py39.py` lines 268-318
```
268:     from docs_src.security.tutorial005_an_py39 import fake_users_db, verify_password
269: 
270:     assert verify_password("secret", fake_users_db["johndoe"]["hashed_password"])
271: 
272: 
273: @needs_py39
274: def test_get_password_hash():
275:     from docs_src.security.tutorial005_an_py39 import get_password_hash
276: 
277:     assert get_password_hash("secretalice")
278: 
279: 
280: @needs_py39
281: def test_create_access_token():
282:     from docs_src.security.tutorial005_an_py39 import create_access_token
283: 
284:     access_token = create_access_token(data={"data": "foo"})
285:     assert access_token
286: 
287: 
288: @needs_py39
289: def test_token_no_sub(client: TestClient):
290:     response = client.get(
291:         "/users/me",
292:         headers={
293:             "Authorization": "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjoiZm9vIn0.9ynBhuYb4e6aW3oJr_K_TBgwcMTDpRToQIE25L57rOE"
294:         },
295:     )
296:     assert response.status_code == 401, response.text
297:     assert response.json() == {"detail": "Could not validate credentials"}
298:     assert response.headers["WWW-Authenticate"] == 'Bearer scope="me"'
299: 
300: 
301: @needs_py39
302: def test_token_no_username(client: TestClient):
303:     response = client.get(
304:         "/users/me",
305:         headers={
306:             "Authorization": "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJmb28ifQ.NnExK_dlNAYyzACrXtXDrcWOgGY2JuPbI4eDaHdfK5Y"
307:         },
308:     )
309:     assert response.status_code == 401, response.text
310:     assert response.json() == {"detail": "Could not validate credentials"}
311:     assert response.headers["WWW-Authenticate"] == 'Bearer scope="me"'
312: 
313: 
314: @needs_py39
315: def test_token_no_scope(client: TestClient):
316:     access_token = get_access_token(client=client)
317:     response = client.get(
318:         "/users/me", headers={"Authorization": f"Bearer {access_token}"}
```

- **A** (claude-opus-5-5): `false_positive` — rule: R3 Path context — The flagged JWT at line 293 is a hardcoded bearer token inside a test file under `tests/`, used to assert a 401 rejection. The mechanism is present only in test code, so it is not a shipped credential.
- **B** (gpt-6.1-sol): `false_positive` — rule: R3 Path context — The JWT appears in a test file under tests/ and is used in a test asserting that invalid credentials return HTTP 401.

**Proposed label: `false_positive`**

Your decision: [ ] confirm  [ ] change to ______  note: ____

---

## f1fc5c0ac3aa

- tool_signal: security.semgrep.actions-mutable-ref rule scout.security.actions.mutable-ref at .github/workflows/ci.yml:176
- path: `.github/workflows/ci.yml`

`.github/workflows/ci.yml` lines 151-184
```
151:       - name: Run Tests
152:         run: |
153:           mkdir -p ~/.ssh && echo -e "${SSH_KEY//_/\\n}" > ~/.ssh/id_ecdsa && chmod og-rwx ~/.ssh/id_ecdsa
154: 
155:           # short sha is enough?
156:           short_sha=$(git rev-parse --short HEAD)
157: 
158:           # The environment is fresh, so there's no point in keeping accepting and adding the key.
159:           rsync -arz -e "ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null" --progress --delete --exclude '.git' . "$CI_USER"@ci-s390x.caddyserver.com:/var/tmp/"$short_sha"
160:           ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -t "$CI_USER"@ci-s390x.caddyserver.com "cd /var/tmp/$short_sha; go version; go env; printf "\n\n";CGO_ENABLED=0 go test -v ./..."
161:           test_result=$?
162: 
163:           # There's no need leaving the files around
164:           ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null "$CI_USER"@ci-s390x.caddyserver.com "rm -rf /var/tmp/'$short_sha'"
165: 
166:           echo "Test exit code: $test_result"
167:           exit $test_result
168:         env:
169:           SSH_KEY: ${{ secrets.S390X_SSH_KEY }}
170:           CI_USER: ${{ secrets.CI_USER }}
171: 
172:   goreleaser-check:
173:     runs-on: ubuntu-latest
174:     steps:
175:       - name: checkout
176:         uses: actions/checkout@v3
177:       
178:       - uses: goreleaser/goreleaser-action@v4
179:         with:
180:           version: latest
181:           args: check
182:         env:
183:           TAG: ${{ steps.vars.outputs.version_tag }}
184: 
```

- **A** (claude-opus-5-5): `real_defect` — rule: R1 Presence — Line 176 has `uses: actions/checkout@v3`. That is a mutable tag, not a pinned commit SHA, in a real CI workflow, and nothing in the excerpt guards or neutralises it.
- **B** (gpt-6.1-sol): `real_defect` — rule: R1 Presence — Line 176 uses actions/checkout@v3, a mutable tag rather than an immutable commit SHA. The workflow excerpt shows no pin or other construct neutralising the mutable reference.

**Proposed label: `real_defect`**

Your decision: [ ] confirm  [ ] change to ______  note: ____

---

## f7ede452a928

- tool_signal: security.gitleaks.default-credentials rule generic-api-key at axum/src/extract/ws.rs:834
- path: `axum/src/extract/ws.rs`

`axum/src/extract/ws.rs` lines 809-859
```
809: }
810: 
811: #[cfg(test)]
812: mod tests {
813:     use super::*;
814:     use crate::{body::Body, routing::get, Router};
815:     use http::{Request, Version};
816:     use tower::ServiceExt;
817: 
818:     #[crate::test]
819:     async fn rejects_http_1_0_requests() {
820:         let svc = get(|ws: Result<WebSocketUpgrade, WebSocketUpgradeRejection>| {
821:             let rejection = ws.unwrap_err();
822:             assert!(matches!(
823:                 rejection,
824:                 WebSocketUpgradeRejection::ConnectionNotUpgradable(_)
825:             ));
826:             std::future::ready(())
827:         });
828: 
829:         let req = Request::builder()
830:             .version(Version::HTTP_10)
831:             .method(Method::GET)
832:             .header("upgrade", "websocket")
833:             .header("connection", "Upgrade")
834:             .header("sec-websocket-key", "6D69KGBOr4Re+Nj6zx9aQA==")
835:             .header("sec-websocket-version", "13")
836:             .body(Body::empty())
837:             .unwrap();
838: 
839:         let res = svc.oneshot(req).await.unwrap();
840: 
841:         assert_eq!(res.status(), StatusCode::OK);
842:     }
843: 
844:     #[allow(dead_code)]
845:     fn default_on_failed_upgrade() {
846:         async fn handler(ws: WebSocketUpgrade) -> Response {
847:             ws.on_upgrade(|_| async {})
848:         }
849:         let _: Router = Router::new().route("/", get(handler));
850:     }
851: 
852:     #[allow(dead_code)]
853:     fn on_failed_upgrade() {
854:         async fn handler(ws: WebSocketUpgrade) -> Response {
855:             ws.on_failed_upgrade(|_error: Error| println!("oops!"))
856:                 .on_upgrade(|_| async {})
857:         }
858:         let _: Router = Router::new().route("/", get(handler));
859:     }
```

- **A** (claude-opus-5-5): `false_positive` — rule: R1 Presence / false_positive criteria (mechanism absent; in test code) — The flagged string at line 834 is a `sec-websocket-key` handshake nonce, not a credential, so no API key is present. It also sits inside a `#[cfg(test)] mod tests` block that is compiled only for tests and never shipped.
- **B** (gpt-6.1-sol): `false_positive` — rule: R3 Path context; test code exclusion — The flagged value is inside a #[cfg(test)] module and a test function. The excerpt therefore shows it only in test code.

**Proposed label: `false_positive`**

Your decision: [ ] confirm  [ ] change to ______  note: ____

---

## go-security-release-mutable-actions-0db29e2c

- tool_signal: semgrep rule SECURITY002 and related: GitHub Action referenced by version tag or branch instead of commit SHA
- path: `.github/workflows/release.yml`

`.github/workflows/release.yml` lines 1-125
```
1: name: Release
2: 
3: on:
4:   push:
5:     tags:
6:       - 'v*.*.*'
7: 
8: jobs:
9:   release:
10:     name: Release
11:     strategy:
12:       matrix:
13:         os: [ ubuntu-latest ]
14:         go: [ '1.20' ]
15: 
16:         include:
17:         # Set the minimum Go patch version for the given Go minor
18:         # Usable via ${{ matrix.GO_SEMVER }}
19:         - go: '1.20'
20:           GO_SEMVER: '~1.20.0'
21: 
22:     runs-on: ${{ matrix.os }}
23:     # https://github.com/sigstore/cosign/issues/1258#issuecomment-1002251233
24:     # https://docs.github.com/en/actions/deployment/security-hardening-your-deployments/about-security-hardening-with-openid-connect#adding-permissions-settings
25:     permissions:
26:       id-token: write
27:       # https://docs.github.com/en/rest/overview/permissions-required-for-github-apps#permission-on-contents
28:       # "Releases" is part of `contents`, so it needs the `write`
29:       contents: write
30: 
31:     steps:
32:     - name: Install Go
33:       uses: actions/setup-go@v3
34:       with:
35:         go-version: ${{ matrix.GO_SEMVER }}
36:         check-latest: true
37: 
38:     - name: Checkout code
39:       uses: actions/checkout@v3
40:       with:
41:         fetch-depth: 0
42: 
43:     # Force fetch upstream tags -- because 65 minutes
44:     # tl;dr: actions/checkout@v3 runs this line:
45:     #   git -c protocol.version=2 fetch --no-tags --prune --progress --no-recurse-submodules --depth=1 origin +ebc278ec98bb24f2852b61fde2a9bf2e3d83818b:refs/tags/
46:     # which makes its own local lightweight tag, losing all the annotations in the process. Our earlier script ran:
47:     #   git fetch --prune --unshallow
48:     # which doesn't overwrite that tag because that would be destructive.
49:     # Credit to @francislavoie for the investigation.
50:     # https://github.com/actions/checkout/issues/290#issuecomment-680260080
51:     - name: Force fetch upstream tags
52:       run: git fetch --tags --force
53: 
54:     # https://github.community/t5/GitHub-Actions/How-to-get-just-the-tag-name/m-p/32167/highlight/true#M1027
55:     - name: Print Go version and environment
56:       id: vars
57:       run: |
58:         printf "Using go at: $(which go)\n"
59:         printf "Go version: $(go version)\n"
60:         printf "\n\nGo environment:\n\n"
61:         go env
62:         printf "\n\nSystem environment:\n\n"
63:         env
64:         echo "version_tag=${GITHUB_REF/refs\/tags\//}" >> $GITHUB_OUTPUT
65:         echo "short_sha=$(git rev-parse --short HEAD)" >> $GITHUB_OUTPUT
66: 
67:         # Add "pip install" CLI tools to PATH
68:         echo ~/.local/bin >> $GITHUB_PATH
69: 
70:         # Parse semver
71:         TAG=${GITHUB_REF/refs\/tags\//}
72:         SEMVER_RE='[^0-9]*\([0-9]*\)[.]\([0-9]*\)[.]\([0-9]*\)\([0-9A-Za-z\.-]*\)'
73:         TAG_MAJOR=`echo ${TAG#v} | sed -e "s#$SEMVER_RE#\1#"`
74:         TAG_MINOR=`echo ${TAG#v} | sed -e "s#$SEMVER_RE#\2#"`
75:         TAG_PATCH=`echo ${TAG#v} | sed -e "s#$SEMVER_RE#\3#"`
76:         TAG_SPECIAL=`echo ${TAG#v} | sed -e "s#$SEMVER_RE#\4#"`
77:         echo "tag_major=${TAG_MAJOR}" >> $GITHUB_OUTPUT
78:         echo "tag_minor=${TAG_MINOR}" >> $GITHUB_OUTPUT
79:         echo "tag_patch=${TAG_PATCH}" >> $GITHUB_OUTPUT
80:         echo "tag_special=${TAG_SPECIAL}" >> $GITHUB_OUTPUT
81: 
82:     # Cloudsmith CLI tooling for pushing releases
83:     # See https://help.cloudsmith.io/docs/cli
84:     - name: Install Cloudsmith CLI
85:       run: pip install --upgrade cloudsmith-cli
86: 
87:     - name: Validate commits and tag signatures
88:       run: |
89:         
90:         # Import Matt Holt's key
91:         curl 'https://github.com/mholt.gpg' | gpg --import
92: 
93:         echo "Verifying the tag: ${{ steps.vars.outputs.version_tag }}"
94:         # tags are only accepted if signed by Matt's key
95:         git verify-tag "${{ steps.vars.outputs.version_tag }}" || exit 1
96: 
97:     - name: Cache the build cache
98:       uses: actions/cache@v3
99:       with:
100:         # In order:
101:         # * Module download cache
102:         # * Build cache (Linux)
103:         path: |
104:           ~/go/pkg/mod
105:           ~/.cache/go-build
106:         key: ${{ runner.os }}-go${{ matrix.go }}-release-${{ hashFiles('**/go.sum') }}
107:         restore-keys: |
108:           ${{ runner.os }}-go${{ matrix.go }}-release
109:     - name: Install Cosign
110:       uses: sigstore/cosign-installer@main
111:     - name: Cosign version
112:       run: cosign version
113:     - name: Install Syft
114:       uses: anchore/sbom-action/download-syft@main
115:     - name: Syft version
116:       run: syft version
117:     # GoReleaser will take care of publishing those artifacts into the release
118:     - name: Run GoReleaser
119:       uses: goreleaser/goreleaser-action@v4
120:       with:
121:         version: latest
122:         args: release --rm-dist --timeout 60m
123:       env:
124:         GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
125:         TAG: ${{ steps.vars.outputs.version_tag }}
```

- **A** (claude-opus-5-5): `real_defect` — rule: R1 Presence (no R2/R3 guard applies) — The release workflow, which is shipped CI config rather than a test or fixture, references actions by mutable tag or branch (e.g. actions/checkout@v3, sigstore/cosign-installer@main, anchore/sbom-action/download-syft@main) with no SHA pin. Nothing in the excerpt neutralises this; the tag signature check covers the repo tag, not the third-party actions.
- **B** (gpt-6.1-sol): `real_defect` — rule: R1 Presence — The release workflow visibly references actions using version tags such as @v3 and @v4 and branches such as @main instead of commit SHAs. No visible guard neutralises those mutable action references; the tag signature check verifies the release tag.

**Proposed label: `real_defect`**

Your decision: [ ] confirm  [ ] change to ______  note: ____

---

## java-architecture-fixture-pom-parse-8ed3c66

- tool_signal: architecture dependency analyzer parsed platform-tests/src/test/resources/listeners/uidtracking/maven/pom.xml as a dependency manifest
- path: `platform-tests/src/test/resources/listeners/uidtracking/maven/pom.xml`

`platform-tests/src/test/resources/listeners/uidtracking/maven/pom.xml` lines 1-2
```
1: <!-- This file is used in tests for the UniqueIdTrackingListener -->
2: 
```

- **A** (claude-opus-5-5): `false_positive` — rule: R3 Path context — The pom.xml is under `src/test/resources/` and its first line states it is used in tests for the UniqueIdTrackingListener, so it is a test fixture rather than a shipped dependency manifest.
- **B** (gpt-6.1-sol): `false_positive` — rule: R3 Path context — The path is under test resources, and the excerpt explicitly states that the file is used in tests for UniqueIdTrackingListener.

**Proposed label: `false_positive`**

Your decision: [ ] confirm  [ ] change to ______  note: ____

---

## java-quality-snapshot-duplication-8ed3c66

- tool_signal: jscpd: duplicate code block between Skip-skipWithSingleLineReason-summary-unicode.out.txt and ...-summary-ascii.out.txt lines 2-14
- path: `platform-tests/src/test/resources/console/details/skip/Skip-skipWithSingleLineReason-summary-unicode.out.txt`

`platform-tests/src/test/resources/console/details/skip/Skip-skipWithSingleLineReason-summary-unicode.out.txt` lines 1-15
```
1: 
2: Test run finished after [\d]+ ms
3: [         2 containers found      ]
4: [         0 containers skipped    ]
5: [         2 containers started    ]
6: [         0 containers aborted    ]
7: [         2 containers successful ]
8: [         0 containers failed     ]
9: [         1 tests found           ]
10: [         1 tests skipped         ]
11: [         0 tests started         ]
12: [         0 tests aborted         ]
13: [         0 tests successful      ]
14: [         0 tests failed          ]
15: 
```

`platform-tests/src/test/resources/console/details/skip/Skip-skipWithSingleLineReason-summary-ascii.out.txt` lines 1-15
```
1: 
2: Test run finished after [\d]+ ms
3: [         2 containers found      ]
4: [         0 containers skipped    ]
5: [         2 containers started    ]
6: [         0 containers aborted    ]
7: [         2 containers successful ]
8: [         0 containers failed     ]
9: [         1 tests found           ]
10: [         1 tests skipped         ]
11: [         0 tests started         ]
12: [         0 tests aborted         ]
13: [         0 tests successful      ]
14: [         0 tests failed          ]
15: 
```

- **A** (claude-opus-5-5): `false_positive` — rule: R3 Path context (also R5 Metrics) — Both files are expected-output fixtures under `src/test/resources`, so the duplication exists only in test resource files. The duplicated block is declarative expected console output, not hand-written logic, so it does not reflect a real maintainability cost.
- **B** (gpt-6.1-sol): `false_positive` — rule: R3 Path context — The duplicated text appears in test resource files under src/test/resources. It is expected console output, not shipped hand-written logic.

**Proposed label: `false_positive`**

Your decision: [ ] confirm  [ ] change to ______  note: ____

---

## java-security-unpinned-github-actions-8ed3c66

- tool_signal: semgrep rules SECURITY001-SECURITY030: GitHub Action referenced by version tag instead of commit SHA
- path: `.github/workflows/main.yml`

`.github/workflows/main.yml` lines 1-60
```
1: name: CI
2: 
3: on:
4:   push:
5:     branches:
6:       - main
7:       - 'releases/*'
8:   pull_request:
9:     branches:
10:       - '*'
11: 
12: env:
13:   ORG_GRADLE_PROJECT_enableTestDistribution: true
14:   ORG_GRADLE_PROJECT_junitBuildCacheUsername: ${{ secrets.BUILD_CACHE_USERNAME }}
15:   ORG_GRADLE_PROJECT_junitBuildCachePassword: ${{ secrets.BUILD_CACHE_PASSWORD }}
16:   GRADLE_ENTERPRISE_ACCESS_KEY: ${{ secrets.GRADLE_ENTERPRISE_ACCESS_KEY }}
17: 
18: jobs:
19:   Linux:
20:     runs-on: ubuntu-latest
21:     steps:
22:     - name: Check out repository
23:       uses: actions/checkout@v3
24:       with:
25:         fetch-depth: 1
26:     - name: Install Graphviz
27:       run: |
28:         sudo apt-get update
29:         sudo apt-get install graphviz
30:     - name: Install GraalVM
31:       uses: graalvm/setup-graalvm@v1
32:       with:
33:         version: 'latest'
34:         java-version: '17'
35:         components: 'native-image'
36:         github-token: ${{ secrets.GITHUB_TOKEN }}
37:     - name: Build
38:       uses: ./.github/actions/main-build
39:       with:
40:         arguments: |
41:           -PenableJaCoCo
42:           build
43:           jacocoRootReport
44:           prepareDocsForUploadToGhPages
45:     - name: Upload to Codecov.io
46:       uses: codecov/codecov-action@v3
47: 
48:   Windows:
49:     runs-on: windows-latest
50:     steps:
51:     - name: Check out repository
52:       uses: actions/checkout@v3
53:       with:
54:         fetch-depth: 1
55:     - name: Build
56:       uses: ./.github/actions/main-build
57: 
58:   macOS:
59:     runs-on: macos-latest
60:     steps:
```

`.github/actions/run-gradle/action.yml` lines 1-26
```
1: name: Run Gradle
2: description: Sets up Gradle JDKs and runs Gradle
3: inputs:
4:   arguments:
5:     required: true
6:     description: Gradle arguments
7:     default: build
8: runs:
9:   using: "composite"
10:   steps:
11:     - uses: actions/setup-java@v3
12:       id: setup-gradle-jdk
13:       with:
14:         distribution: temurin
15:         java-version: 17
16:     - uses: gradle/gradle-build-action@v2
17:       env:
18:         JAVA_HOME: ${{ steps.setup-gradle-jdk.outputs.path }}
19:       with:
20:         arguments: |
21:           -Porg.gradle.java.installations.auto-download=false
22:           -PenablePredictiveTestSelection=${{ github.event_name == 'pull_request' }}
23:           "-Dscan.value.GitHub job=${{ github.job }}"
24:           javaToolchains
25:           ${{ inputs.arguments }}
26: 
```

- **A** (claude-opus-5-5): `real_defect` — rule: R1 Presence (no R2/R3/R4 guard applies) — The workflow, which is shipped CI configuration, references third-party actions by mutable version tags (e.g. `actions/checkout@v3`, `graalvm/setup-graalvm@v1`, `codecov/codecov-action@v3`), so the flagged mechanism is literally present. No SHA pin or other neutralising construct appears anywhere in the excerpt.
- **B** (gpt-6.1-sol): `real_defect` — rule: R1 Presence — The workflow visibly references actions/checkout@v3, graalvm/setup-graalvm@v1, and codecov/codecov-action@v3 using version tags rather than commit SHAs. No visible guard or pin neutralises this mechanism.

**Proposed label: `real_defect`**

Your decision: [ ] confirm  [ ] change to ______  note: ____

---

## quality-long-openapi-test-fixture-d666ccb6

- tool_signal: lizard rule COMPLEXITY001: function length 194 lines, cyclomatic complexity 1
- path: `tests/test_generate_unique_id_function.py`

`tests/test_generate_unique_id_function.py` lines 20-75
```
20: 
21: 
22: class Item(BaseModel):
23:     name: str
24:     price: float
25: 
26: 
27: class Message(BaseModel):
28:     title: str
29:     description: str
30: 
31: 
32: def test_top_level_generate_unique_id():
33:     app = FastAPI(generate_unique_id_function=custom_generate_unique_id)
34:     router = APIRouter()
35: 
36:     @app.post("/", response_model=List[Item], responses={404: {"model": List[Message]}})
37:     def post_root(item1: Item, item2: Item):
38:         return item1, item2  # pragma: nocover
39: 
40:     @router.post(
41:         "/router", response_model=List[Item], responses={404: {"model": List[Message]}}
42:     )
43:     def post_router(item1: Item, item2: Item):
44:         return item1, item2  # pragma: nocover
45: 
46:     app.include_router(router)
47:     client = TestClient(app)
48:     response = client.get("/openapi.json")
49:     data = response.json()
50:     assert data == {
51:         "openapi": "3.0.2",
52:         "info": {"title": "FastAPI", "version": "0.1.0"},
53:         "paths": {
54:             "/": {
55:                 "post": {
56:                     "summary": "Post Root",
57:                     "operationId": "foo_post_root",
58:                     "requestBody": {
59:                         "content": {
60:                             "application/json": {
61:                                 "schema": {
62:                                     "$ref": "#/components/schemas/Body_foo_post_root"
63:                                 }
64:                             }
65:                         },
66:                         "required": True,
67:                     },
68:                     "responses": {
69:                         "200": {
70:                             "description": "Successful Response",
71:                             "content": {
72:                                 "application/json": {
73:                                     "schema": {
74:                                         "title": "Response Foo Post Root",
75:                                         "type": "array",
```

- **A** (claude-opus-5-5): `false_positive` — rule: R3 Path context (also R5 Metrics) — The file is in `tests/` and the long function is a test whose length comes from a large literal OpenAPI dict in an assertion, with cyclomatic complexity 1. That is repetitive declarative content, not hand-written shipped logic.
- **B** (gpt-6.1-sol): `false_positive` — rule: R3 Path context — The finding is in tests/test_generate_unique_id_function.py, and the excerpt shows a test function. Findings confined to test code are false positives under R3.

**Proposed label: `false_positive`**

Your decision: [ ] confirm  [ ] change to ______  note: ____

---

## rust-architecture-benchmark-blocking-false-positive-562fa206

- tool_signal: architecture scalability analyzer rule ARCHITECTURE001 at axum/benches/benches.rs lines 12-25 region: blocking call (child.wait / synchronous I/O)
- path: `axum/benches/benches.rs`

`axum/benches/benches.rs` lines 175-225
```
175:             cmd.args(["--duration", "10s"]);
176:         }
177: 
178:         if let Some(method) = self.method {
179:             cmd.args(["--method", method]);
180:         }
181: 
182:         for (key, value) in self.headers.into_iter().flatten() {
183:             cmd.arg("--header");
184:             cmd.arg(format!("{key}: {value}"));
185:         }
186: 
187:         if let Some(body) = self.body {
188:             cmd.args(["--body", body]);
189:         }
190: 
191:         eprintln!("Running {:?} benchmark", self.name);
192: 
193:         // indent output from `rewrk` so its easier to read when running multiple benchmarks
194:         let mut child = cmd.spawn().unwrap();
195:         let stdout = child.stdout.take().unwrap();
196:         let stdout = std::io::BufReader::new(stdout);
197:         for line in stdout.lines() {
198:             let line = line.unwrap();
199:             println!("  {line}");
200:         }
201: 
202:         let status = child.wait().unwrap();
203: 
204:         if !status.success() {
205:             eprintln!("`rewrk` command failed");
206:             std::process::exit(status.code().unwrap());
207:         }
208:     }
209: }
210: 
211: fn install_rewrk() {
212:     println!("installing rewrk");
213:     let mut cmd = Command::new("cargo");
214:     cmd.args([
215:         "install",
216:         "rewrk",
217:         "--git",
218:         "https://github.com/ChillFish8/rewrk.git",
219:     ]);
220:     let status = cmd
221:         .status()
222:         .unwrap_or_else(|_| panic!("failed to install rewrk"));
223:     if !status.success() {
224:         panic!("failed to install rewrk");
225:     }
```

- **A** (claude-opus-5-5): `false_positive` — rule: R3 Path context — The file `axum/benches/benches.rs` is a benchmark harness, a test-like file that is not shipped library code, and the excerpt shows the blocking `child.wait()` and `cmd.status()` only in its driver for the `rewrk` tool. Synchronous process waiting in a benchmark runner is not a scalability issue in shipped code, though the exact flagged lines 12-25 are outside the excerpt.
- **B** (gpt-6.1-sol): `false_positive` — rule: R3 Path context — The blocking calls appear in axum/benches/benches.rs, a benchmark file rather than shipped runtime code. The benchmark context makes this false_positive.

**Proposed label: `false_positive`**

Your decision: [ ] confirm  [ ] change to ______  note: ____

---

## rust-performance-websocket-demo-resource-false-positive-562fa206

- tool_signal: frontend performance analyzer rules PERFORMANCE001-PERFORMANCE004: timers / event listeners without cleanup (resource leak)
- path: `examples/websockets/assets/script.js`

`examples/websockets/assets/script.js` lines 1-25
```
1: const socket = new WebSocket('ws://localhost:3000/ws');
2: 
3: socket.addEventListener('open', function (event) {
4:     socket.send('Hello Server!');
5: });
6: 
7: socket.addEventListener('message', function (event) {
8:     console.log('Message from server ', event.data);
9: });
10: 
11: 
12: setTimeout(() => {
13:     const obj = { hello: "world" };
14:     const blob = new Blob([JSON.stringify(obj, null, 2)], {
15:       type: "application/json",
16:     });
17:     console.log("Sending blob over websocket");
18:     socket.send(blob);
19: }, 1000);
20: 
21: setTimeout(() => {
22:     socket.send('About done here...');
23:     console.log("Sending close over websocket");
24:     socket.close(3000, "Crash and Burn!");
25: }, 3000);
```

- **A** (claude-opus-5-5): `false_positive` — rule: R3 Path context — The timers and event listeners are present, but only in `examples/websockets/assets/script.js`, which is clearly an example file. R3 makes the finding false_positive regardless of the missing cleanup.
- **B** (gpt-6.1-sol): `false_positive` — rule: R3 Path context — The timers and event listeners occur in examples/websockets/assets/script.js, which is clearly an example file.

**Proposed label: `false_positive`**

Your decision: [ ] confirm  [ ] change to ______  note: ____

---

## rust-quality-documented-query-overlap-false-positive-562fa206

- tool_signal: jscpd rules QUALITY209-QUALITY210: duplicate code between axum/src/extract/query.rs and axum-extra/src/extract/query.rs
- path: `axum/src/extract/query.rs`

`axum/src/extract/query.rs` lines 1-60
```
1: use super::{rejection::*, FromRequestParts};
2: use async_trait::async_trait;
3: use http::request::Parts;
4: use serde::de::DeserializeOwned;
5: use std::ops::Deref;
6: 
7: /// Extractor that deserializes query strings into some type.
8: ///
9: /// `T` is expected to implement [`serde::Deserialize`].
10: ///
11: /// # Example
12: ///
13: /// ```rust,no_run
14: /// use axum::{
15: ///     extract::Query,
16: ///     routing::get,
17: ///     Router,
18: /// };
19: /// use serde::Deserialize;
20: ///
21: /// #[derive(Deserialize)]
22: /// struct Pagination {
23: ///     page: usize,
24: ///     per_page: usize,
25: /// }
26: ///
27: /// // This will parse query strings like `?page=2&per_page=30` into `Pagination`
28: /// // structs.
29: /// async fn list_things(pagination: Query<Pagination>) {
30: ///     let pagination: Pagination = pagination.0;
31: ///
32: ///     // ...
33: /// }
34: ///
35: /// let app = Router::new().route("/list_things", get(list_things));
36: /// # async {
37: /// # axum::Server::bind(&"".parse().unwrap()).serve(app.into_make_service()).await.unwrap();
38: /// # };
39: /// ```
40: ///
41: /// If the query string cannot be parsed it will reject the request with a `400
42: /// Bad Request` response.
43: ///
44: /// For handling values being empty vs missing see the [query-params-with-empty-strings][example]
45: /// example.
46: ///
47: /// [example]: https://github.com/tokio-rs/axum/blob/main/examples/query-params-with-empty-strings/src/main.rs
48: #[cfg_attr(docsrs, doc(cfg(feature = "query")))]
49: #[derive(Debug, Clone, Copy, Default)]
50: pub struct Query<T>(pub T);
51: 
52: #[async_trait]
53: impl<T, S> FromRequestParts<S> for Query<T>
54: where
55:     T: DeserializeOwned,
56:     S: Send + Sync,
57: {
58:     type Rejection = QueryRejection;
59: 
60:     async fn from_request_parts(parts: &mut Parts, _state: &S) -> Result<Self, Self::Rejection> {
```

`axum-extra/src/extract/query.rs` lines 1-65
```
1: use axum::{
2:     async_trait,
3:     extract::FromRequestParts,
4:     response::{IntoResponse, Response},
5:     Error,
6: };
7: use http::{request::Parts, StatusCode};
8: use serde::de::DeserializeOwned;
9: use std::{fmt, ops::Deref};
10: 
11: /// Extractor that deserializes query strings into some type.
12: ///
13: /// `T` is expected to implement [`serde::Deserialize`].
14: ///
15: /// # Differences from `axum::extract::Query`
16: ///
17: /// This extractor uses [`serde_html_form`] under-the-hood which supports multi-value items. These
18: /// are sent by multiple `<input>` attributes of the same name (e.g. checkboxes) and `<select>`s
19: /// with the `multiple` attribute. Those values can be collected into a `Vec` or other sequential
20: /// container.
21: ///
22: /// # Example
23: ///
24: /// ```rust,no_run
25: /// use axum::{routing::get, Router};
26: /// use axum_extra::extract::Query;
27: /// use serde::Deserialize;
28: ///
29: /// #[derive(Deserialize)]
30: /// struct Pagination {
31: ///     page: usize,
32: ///     per_page: usize,
33: /// }
34: ///
35: /// // This will parse query strings like `?page=2&per_page=30` into `Pagination`
36: /// // structs.
37: /// async fn list_things(pagination: Query<Pagination>) {
38: ///     let pagination: Pagination = pagination.0;
39: ///
40: ///     // ...
41: /// }
42: ///
43: /// let app = Router::new().route("/list_things", get(list_things));
44: /// # async {
45: /// # axum::Server::bind(&"".parse().unwrap()).serve(app.into_make_service()).await.unwrap();
46: /// # };
47: /// ```
48: ///
49: /// If the query string cannot be parsed it will reject the request with a `400
50: /// Bad Request` response.
51: ///
52: /// For handling values being empty vs missing see the [query-params-with-empty-strings][example]
53: /// example.
54: ///
55: /// [example]: https://github.com/tokio-rs/axum/blob/main/examples/query-params-with-empty-strings/src/main.rs
56: #[cfg_attr(docsrs, doc(cfg(feature = "query")))]
57: #[derive(Debug, Clone, Copy, Default)]
58: pub struct Query<T>(pub T);
59: 
60: #[async_trait]
61: impl<T, S> FromRequestParts<S> for Query<T>
62: where
63:     T: DeserializeOwned,
64:     S: Send + Sync,
65: {
```

- **A** (claude-opus-5-5): `false_positive` — rule: R5 Metrics: duplication inflated by repetitive declarative content — The visible duplicated region is almost entirely doc-comment text, a derive/struct declaration and a trait impl signature, which is declarative boilerplate rather than duplicated hand-written logic. The axum-extra docs also note it deliberately uses a different deserializer (`serde_html_form`), so the overlap does not reflect a real maintainability cost.
- **B** (gpt-6.1-sol): `false_positive` — rule: R5 Metrics — The visible duplication consists of documentation examples and repetitive type and trait declarations, which inflate the duplication metric without showing duplicated implementation logic.

**Proposed label: `false_positive`**

Your decision: [ ] confirm  [ ] change to ______  note: ____

---

## rust-security-local-websocket-test-false-positive-562fa206

- tool_signal: security analyzer rule SECURITY053: plain ws:// (unencrypted WebSocket) URL
- path: `examples/testing-websockets/src/main.rs`

`examples/testing-websockets/src/main.rs` lines 80-125
```
80: {
81:     while let Some(Ok(msg)) = read.next().await {
82:         if let Message::Text(msg) = msg {
83:             if write
84:                 .send(Message::Text(format!("You said: {msg}")))
85:                 .await
86:                 .is_err()
87:             {
88:                 break;
89:             }
90:         }
91:     }
92: }
93: 
94: #[cfg(test)]
95: mod tests {
96:     use super::*;
97:     use std::net::Ipv4Addr;
98:     use tokio_tungstenite::tungstenite;
99: 
100:     // We can integration test one handler by running the server in a background task and
101:     // connecting to it like any other client would.
102:     #[tokio::test]
103:     async fn integration_test() {
104:         let server = axum::Server::bind(&SocketAddr::from((Ipv4Addr::UNSPECIFIED, 0)))
105:             .serve(app().into_make_service());
106:         let addr = server.local_addr();
107:         tokio::spawn(server);
108: 
109:         let (mut socket, _response) =
110:             tokio_tungstenite::connect_async(format!("ws://{addr}/integration-testable"))
111:                 .await
112:                 .unwrap();
113: 
114:         socket
115:             .send(tungstenite::Message::text("foo"))
116:             .await
117:             .unwrap();
118: 
119:         let msg = match socket.next().await.unwrap().unwrap() {
120:             tungstenite::Message::Text(msg) => msg,
121:             other => panic!("expected a text message but got {other:?}"),
122:         };
123: 
124:         assert_eq!(msg, "You said: foo");
125:     }
```

- **A** (claude-opus-5-5): `false_positive` — rule: R3 Path context — The ws:// URL appears only inside a #[cfg(test)] integration test, in a file under the examples/ directory. It connects to a local ephemeral test server, so it is not shipped production code.
- **B** (gpt-6.1-sol): `false_positive` — rule: R3 Path context — The ws:// URL occurs in an examples/ file and inside a #[cfg(test)] module's integration test. The flagged mechanism is present only in example/test code.

**Proposed label: `false_positive`**

Your decision: [ ] confirm  [ ] change to ______  note: ____

---

## rust-security-mutable-action-references-562fa206

- tool_signal: semgrep rules SECURITY001-SECURITY049: GitHub Action referenced by branch/version tag instead of commit SHA
- path: `.github/workflows/CI.yml`

`.github/workflows/CI.yml` lines 1-80
```
1: name: CI
2: 
3: env:
4:   CARGO_TERM_COLOR: always
5:   MSRV: '1.60'
6: 
7: on:
8:   push:
9:     branches:
10:     - main
11:   pull_request: {}
12: 
13: jobs:
14:   check:
15:     runs-on: ubuntu-latest
16:     steps:
17:     - uses: actions/checkout@master
18:     - uses: arduino/setup-protoc@v1
19:       with:
20:         repo-token: ${{ secrets.GITHUB_TOKEN }}
21:     - uses: dtolnay/rust-toolchain@stable
22:       with:
23:         toolchain: beta
24:         override: true
25:         profile: minimal
26:         components: clippy, rustfmt
27:     - uses: Swatinem/rust-cache@v2
28:     - name: Check
29:       uses: actions-rs/cargo@v1
30:       with:
31:         command: clippy
32:         args: --workspace --all-targets --all-features
33:     - name: rustfmt
34:       uses: actions-rs/cargo@v1
35:       with:
36:         command: fmt
37:         args: -- --check
38: 
39:   check-docs:
40:     runs-on: ubuntu-latest
41:     steps:
42:     - uses: actions/checkout@master
43:     - uses: dtolnay/rust-toolchain@stable
44:       with:
45:         toolchain: stable
46:         override: true
47:         profile: minimal
48:     - uses: Swatinem/rust-cache@v1
49:     - name: cargo doc
50:       env:
51:         RUSTDOCFLAGS: "-D broken-intra-doc-links"
52:       run: cargo doc --all-features --no-deps
53: 
54:   cargo-hack:
55:     runs-on: ubuntu-latest
56:     steps:
57:     - uses: actions/checkout@master
58:     - uses: arduino/setup-protoc@v1
59:     - uses: dtolnay/rust-toolchain@stable
60:       with:
61:         toolchain: stable
62:         override: true
63:         profile: minimal
64:     - uses: Swatinem/rust-cache@v1
65:     - name: Install cargo-hack
66:       run: |
67:         curl -LsSf https://github.com/taiki-e/cargo-hack/releases/latest/download/cargo-hack-x86_64-unknown-linux-gnu.tar.gz | tar xzf - -C ~/.cargo/bin
68:     - name: cargo hack check
69:       run: cargo hack check --each-feature --no-dev-deps --all
70: 
71:   cargo-public-api-crates:
72:     runs-on: ubuntu-latest
73:     strategy:
74:       matrix:
75:         crate: [axum, axum-core, axum-extra, axum-macros]
76:     steps:
77:     - uses: actions/checkout@master
78:     - uses: dtolnay/rust-toolchain@stable
79:       with:
80:         toolchain: nightly
```

- **A** (claude-opus-5-5): `real_defect` — rule: R1: the flagged mechanism appears literally in the excerpt, and no false_positive rule (R2–R5) applies — The shipped CI workflow uses actions by branch or tag rather than commit SHA, for example `actions/checkout@master`, `arduino/setup-protoc@v1`, `dtolnay/rust-toolchain@stable` and `Swatinem/rust-cache@v2`, and nothing in the excerpt pins or guards them. A `.github/workflows/` file is real project configuration, not a test, fixture, example or generated file, so R3 does not apply.
- **B** (gpt-6.1-sol): `real_defect` — rule: R1 Presence — The workflow visibly references actions by branches and version tags, including actions/checkout@master and arduino/setup-protoc@v1, instead of commit SHAs. No visible guard or pin neutralises the mechanism.

**Proposed label: `real_defect`**

Your decision: [ ] confirm  [ ] change to ______  note: ____

---

## security-auto-public-path-dom-clobbering

- tool_signal: security analyzer structural signal: lib/runtime/AutoPublicPathRuntimeModule.js reads document.currentScript.src and assigns the public path
- path: `lib/runtime/AutoPublicPathRuntimeModule.js`

`lib/runtime/AutoPublicPathRuntimeModule.js` lines 25-70
```
25: 			JavascriptModulesPlugin.getChunkFilenameTemplate(
26: 				this.chunk,
27: 				compilation.outputOptions
28: 			),
29: 			{
30: 				chunk: this.chunk,
31: 				contentHashType: "javascript"
32: 			}
33: 		);
34: 		const undoPath = getUndoPath(chunkName, path, false);
35: 
36: 		return Template.asString([
37: 			"var scriptUrl;",
38: 			scriptType === "module"
39: 				? `if (typeof ${importMetaName}.url === "string") scriptUrl = ${importMetaName}.url`
40: 				: Template.asString([
41: 						`if (${RuntimeGlobals.global}.importScripts) scriptUrl = ${RuntimeGlobals.global}.location + "";`,
42: 						`var document = ${RuntimeGlobals.global}.document;`,
43: 						"if (!scriptUrl && document) {",
44: 						Template.indent([
45: 							`if (document.currentScript)`,
46: 							Template.indent(`scriptUrl = document.currentScript.src;`),
47: 							"if (!scriptUrl) {",
48: 							Template.indent([
49: 								'var scripts = document.getElementsByTagName("script");',
50: 								"if(scripts.length) scriptUrl = scripts[scripts.length - 1].src"
51: 							]),
52: 							"}"
53: 						]),
54: 						"}"
55: 				  ]),
56: 			"// When supporting browsers where an automatic publicPath is not supported you must specify an output.publicPath manually via configuration",
57: 			'// or pass an empty string ("") and set the __webpack_public_path__ variable from your code to use your own logic.',
58: 			'if (!scriptUrl) throw new Error("Automatic publicPath is not supported in this browser");',
59: 			'scriptUrl = scriptUrl.replace(/#.*$/, "").replace(/\\?.*$/, "").replace(/\\/[^\\/]+$/, "/");',
60: 			!undoPath
61: 				? `${RuntimeGlobals.publicPath} = scriptUrl;`
62: 				: `${RuntimeGlobals.publicPath} = scriptUrl + ${JSON.stringify(
63: 						undoPath
64: 				  )};`
65: 		]);
66: 	}
67: }
68: 
69: module.exports = AutoPublicPathRuntimeModule;
70: 
```

- **A** (claude-opus-5-5): `real_defect` — rule: R1 Presence (R2 not satisfied) — Lines 45-46 and 61-64 of shipped, hand-written code under lib/runtime/ literally read document.currentScript.src (falling back to the last script tag's src) and assign it to the public path. The regex on line 59 only strips the fragment, query and filename, so it does not check or constrain where the script URL came from and nothing visible neutralises the mechanism.
- **B** (gpt-6.1-sol): `real_defect` — rule: R1 Presence — The hand-written runtime module emits code that reads document.currentScript.src and assigns the derived URL to the public path. The visible existence checks and URL cleanup do not neutralise that mechanism.

**Proposed label: `real_defect`**

Your decision: [ ] confirm  [ ] change to ______  note: ____

---

## security-python-multipart-redos-test-extra-d666ccb6

- tool_signal: dependency/security analyzer: python-multipart version constraint in pyproject.toml reported as a security finding (CVE-affected range check)
- path: `pyproject.toml`

`pyproject.toml` lines 45-78
```
45:     "pydantic >=1.6.2,!=1.7,!=1.7.1,!=1.7.2,!=1.7.3,!=1.8,!=1.8.1,<2.0.0",
46: ]
47: dynamic = ["version"]
48: 
49: [project.urls]
50: Homepage = "https://github.com/tiangolo/fastapi"
51: Documentation = "https://fastapi.tiangolo.com/"
52: 
53: [project.optional-dependencies]
54: test = [
55:     "pytest >=7.1.3,<8.0.0",
56:     "coverage[toml] >= 6.5.0,< 8.0",
57:     "mypy ==0.982",
58:     "ruff ==0.0.138",
59:     "black == 23.1.0",
60:     "isort >=5.0.6,<6.0.0",
61:     "httpx >=0.23.0,<0.24.0",
62:     "email_validator >=1.1.1,<2.0.0",
63:     # TODO: once removing databases from tutorial, upgrade SQLAlchemy
64:     # probably when including SQLModel
65:     "sqlalchemy >=1.3.18,<1.4.43",
66:     "peewee >=3.13.3,<4.0.0",
67:     "databases[sqlite] >=0.3.2,<0.7.0",
68:     "orjson >=3.2.1,<4.0.0",
69:     "ujson >=4.0.1,!=4.0.2,!=4.1.0,!=4.2.0,!=4.3.0,!=5.0.0,!=5.1.0,<6.0.0",
70:     "python-multipart >=0.0.5,<0.0.7",
71:     "flask >=1.1.2,<3.0.0",
72:     "anyio[trio] >=3.2.1,<4.0.0",
73:     "python-jose[cryptography] >=3.3.0,<4.0.0",
74:     "pyyaml >=5.3.1,<7.0.0",
75:     "passlib[bcrypt] >=1.7.2,<2.0.0",
76: 
77:     # types
78:     "types-ujson ==5.7.0.1",
```

`fastapi/routing.py` lines 175-205
```
175:     response_model_exclude: Optional[Union[SetIntStr, DictIntStrAny]] = None,
176:     response_model_by_alias: bool = True,
177:     response_model_exclude_unset: bool = False,
178:     response_model_exclude_defaults: bool = False,
179:     response_model_exclude_none: bool = False,
180:     dependency_overrides_provider: Optional[Any] = None,
181: ) -> Callable[[Request], Coroutine[Any, Any, Response]]:
182:     assert dependant.call is not None, "dependant.call must be a function"
183:     is_coroutine = asyncio.iscoroutinefunction(dependant.call)
184:     is_body_form = body_field and isinstance(body_field.field_info, params.Form)
185:     if isinstance(response_class, DefaultPlaceholder):
186:         actual_response_class: Type[Response] = response_class.value
187:     else:
188:         actual_response_class = response_class
189: 
190:     async def app(request: Request) -> Response:
191:         try:
192:             body: Any = None
193:             if body_field:
194:                 if is_body_form:
195:                     body = await request.form()
196:                     stack = request.scope.get("fastapi_astack")
197:                     assert isinstance(stack, AsyncExitStack)
198:                     stack.push_async_callback(body.close)
199:                 else:
200:                     body_bytes = await request.body()
201:                     if body_bytes:
202:                         json_body: Any = Undefined
203:                         content_type_value = request.headers.get("content-type")
204:                         if not content_type_value:
205:                             json_body = await request.json()
```

- **A** (claude-opus-5-5): `false_positive` — rule: R3 Path context (test-only scope), which wins under R7 — The flagged `python-multipart >=0.0.5,<0.0.7` constraint appears only in the `[project.optional-dependencies] test` group (line 70), so the vulnerable range is pinned only for the test environment, not for shipped runtime dependencies. The runtime `request.form()` code in routing.py doesn't pin any version, so the excerpt shows no shipped dependency on the affected range.
- **B** (gpt-6.1-sol): `false_positive` — rule: R3 Path context / test-only context — The flagged python-multipart constraint is visibly inside the optional dependency group named test. The routing excerpt does not establish that this test dependency constraint applies to shipped runtime dependencies.

**Proposed label: `false_positive`**

Your decision: [ ] confirm  [ ] change to ______  note: ____

---

## security-unpinned-publish-actions-d666ccb6

- tool_signal: semgrep rules SECURITY022-SECURITY025: GitHub Action referenced by version tag instead of commit SHA
- path: `.github/workflows/publish.yml`

`.github/workflows/publish.yml` lines 1-45
```
1: name: Publish
2: 
3: on:
4:   release:
5:     types:
6:       - created
7: 
8: jobs:
9:   publish:
10:     runs-on: ubuntu-latest
11:     steps:
12:       - name: Dump GitHub context
13:         env:
14:           GITHUB_CONTEXT: ${{ toJson(github) }}
15:         run: echo "$GITHUB_CONTEXT"
16:       - uses: actions/checkout@v3
17:       - name: Set up Python
18:         uses: actions/setup-python@v4
19:         with:
20:           python-version: "3.7"
21:           cache: "pip"
22:           cache-dependency-path: pyproject.toml
23:       - uses: actions/cache@v3
24:         id: cache
25:         with:
26:           path: ${{ env.pythonLocation }}
27:           key: ${{ runner.os }}-python-${{ env.pythonLocation }}-${{ hashFiles('pyproject.toml') }}-publish
28:       - name: Install build dependencies
29:         if: steps.cache.outputs.cache-hit != 'true'
30:         run: pip install build
31:       - name: Build distribution
32:         run: python -m build
33:       - name: Publish
34:         uses: pypa/gh-action-pypi-publish@v1.6.4
35:         with:
36:           password: ${{ secrets.PYPI_API_TOKEN }}
37:       - name: Dump GitHub context
38:         env:
39:           GITHUB_CONTEXT: ${{ toJson(github) }}
40:         run: echo "$GITHUB_CONTEXT"
41:       # - name: Notify
42:       #   env:
43:       #     GITTER_TOKEN: ${{ secrets.GITTER_TOKEN }}
44:       #     GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
45:       #     TAG: ${{ github.event.release.name }}
```

- **A** (claude-opus-5-5): `real_defect` — rule: R1 Presence / R2 unguarded — Lines 16, 18, 23 and 34 literally reference actions by mutable version tags (e.g. actions/checkout@v3, pypa/gh-action-pypi-publish@v1.6.4) rather than commit SHAs, in a shipped CI workflow that is not a test/fixture/docs path. Nothing in the excerpt pins or otherwise neutralises these references.
- **B** (gpt-6.1-sol): `real_defect` — rule: R1 Presence — The publish workflow visibly references actions using version tags, including actions/checkout@v3 and pypa/gh-action-pypi-publish@v1.6.4, instead of commit SHAs. No visible construct guards or neutralises these references.

**Proposed label: `real_defect`**

Your decision: [ ] confirm  [ ] change to ______  note: ____

---

## typescript-architecture-loopback-url-676c98ca

- tool_signal: architecture scalability analyzer rule ARCHITECTURE001 at packages/core/nest-application.ts:361: hardcoded configuration value
- path: `packages/core/nest-application.ts`

`packages/core/nest-application.ts` lines 340-380
```
340:       resolve(this.formatAddress(address));
341:     });
342:   }
343: 
344:   private formatAddress(address: any): string {
345:     if (isString(address)) {
346:       if (platform() === 'win32') {
347:         return address;
348:       }
349:       const basePath = encodeURIComponent(address);
350:       return `${this.getProtocol()}+unix://${basePath}`;
351:     }
352: 
353:     let host = this.host();
354:     if (address && address.family === 'IPv6') {
355:       if (host === '::') {
356:         host = '[::1]';
357:       } else {
358:         host = `[${host}]`;
359:       }
360:     } else if (host === '0.0.0.0') {
361:       host = '127.0.0.1';
362:     }
363: 
364:     return `${this.getProtocol()}://${host}:${address.port}`;
365:   }
366: 
367:   public setGlobalPrefix(prefix: string, options?: GlobalPrefixOptions): this {
368:     this.config.setGlobalPrefix(prefix);
369:     if (options) {
370:       const exclude = options?.exclude
371:         ? mapToExcludeRoute(options.exclude)
372:         : [];
373:       this.config.setGlobalPrefixOptions({
374:         ...options,
375:         exclude,
376:       });
377:     }
378:     return this;
379:   }
380: 
```

- **A** (claude-opus-5-5): `false_positive` — rule: R2 Guarded vs unguarded — The literal '127.0.0.1' is not a hardcoded configuration value. It is used only inside the visible `host === '0.0.0.0'` check, to turn the configured wildcard bind address into a loopback URL for display, and otherwise the host comes from `this.host()`.
- **B** (gpt-6.1-sol): `false_positive` — rule: R2 Guarded vs unguarded — The literal '127.0.0.1' is guarded by a check for host === '0.0.0.0' inside address formatting. It normalises the wildcard address for the returned URL rather than setting a hardcoded application configuration.

**Proposed label: `false_positive`**

Your decision: [ ] confirm  [ ] change to ______  note: ____

---

## typescript-performance-sponsor-images-676c98ca

- tool_signal: frontend performance analyzer rule PERFORMANCE012 at Readme.md:63: unoptimized image format
- path: `Readme.md`

`Readme.md` lines 55-90
```
55: 
56: Nest is an MIT-licensed open source project. It can grow thanks to the sponsors and support from the amazing backers. If you'd like to join them, please [read more here](https://docs.nestjs.com/support).
57: 
58: #### Principal Sponsors
59: <table style="text-align:center;"><tr>
60:   <td><a href="https://trilon.io" target="_blank"><img src="https://nestjs.com/img/trilon.svg" width="200" valign="middle" /></a></td>
61:   <td><a href="https://www.redhat.com" target="_blank"><img src="https://nestjs.com/img/red-hat-logo.svg" width="200" valign="middle" /></a></td>
62: <td>
63: <a href="https://github.com/Sanofi-IADC" target="_blank"><img src="https://docs.nestjs.com/assets/sponsors/sanofi.png" width="180" valign="middle" /></a></td>
64: <td>
65: <a href="https://nx.dev" target="_blank"><img src="https://nestjs.com/img/nx-logo.png" height="45" valign="middle" /></a></td>
66:     <td>
67: <a href="https://valor-software.com/" target="_blank"><img src="https://docs.nestjs.com/assets/sponsors/valor-software.png" width="170" valign="middle" /></a></td><td>
68: <a href="https://amplication.com/" target="_blank"><img src="https://nestjs.com/img/amplication-logo.svg" width="190" valign="middle" /></a></td>
69: </tr></table>
70: 
71: #### Gold Sponsors
72: 
73: <table style="text-align:center;"><tr><td>
74: <a href="https://weld.app/" target="_blank"><img src="https://nestjs.com/img/weld-logo.svg" width="140" valign="middle" /></a></td>
75: <td>
76: <a href="https://intrinsic.ventures/" target="_blank"><img src="https://nestjs.com/img/intrinisic-logo.png" width="210" valign="middle" /></a></td>
77:  <td>
78: <a href="https://jetbrains.com/" target="_blank"><img src="https://nestjs.com/img/jetbrains-logo.svg" width="110" valign="middle" /></a></td><td>
79: <a href="https://snyk.co/nestjs" target="_blank"><img src="https://nestjs.com/img/snyk-logo-black.png" width="185" valign="middle" /></a></td><td>
80: <a href="https://fuseautotech.com/" target="_blank"><img src="https://nestjs.com/img/fuse-logo.svg" width="105" valign="middle" /></a></td>
81: <td>
82: <a href="https://ridicorp.com/career/" target="_blank"><img src="https://nestjs.com/img/ridi-logo.svg" width="105" valign="middle" /></a></td><td>
83: <a href="https://www.movavi.com/imovie-for-windows.html" target="_blank"><img src="https://nestjs.com/img/movavi-logo.svg" width="105" valign="middle" /></a></td></</tr></table>
84: 
85: #### Silver Sponsors
86: 
87: <table style="text-align:center;"><tr><td>
88: <a href="https://neoteric.eu/" target="_blank"><img src="https://nestjs.com/img/neoteric-cut.png" width="120" valign="middle" /></a> </td><td>
89:   <a href="http://gojob.com" target="_blank"><img src="http://nestjs.com/img/gojob-logo.png" valign="middle" width="100" /></a> </td><td>
90: <a href="https://www.myleodsc.com/" target="_blank"><img src="https://nestjs.com/img/myleo-logo.png" width="180" valign="middle" /></td><td>
```

- **A** (claude-opus-5-5): `false_positive` — rule: R3 Path context — The flagged PNG at line 63 is in Readme.md, which is a documentation file rather than shipped frontend code. A sponsor logo in a README carries no runtime cost for the application.
- **B** (gpt-6.1-sol): `false_positive` — rule: R3 Path context — The flagged PNG image appears only in Readme.md, a documentation file. Documentation findings qualify as false_positive under R3.

**Proposed label: `false_positive`**

Your decision: [ ] confirm  [ ] change to ______  note: ____

---

## typescript-security-mutable-github-actions-676c98ca

- tool_signal: semgrep rules SECURITY003-SECURITY006: GitHub Action referenced by version tag instead of commit SHA
- path: `.github/workflows/codeql-analysis.yml`

`.github/workflows/codeql-analysis.yml` lines 1-62
```
1: name: "CodeQL"
2: 
3: on:
4:   push:
5:     branches: [master, ]
6:   pull_request:
7:     # The branches below must be a subset of the branches above
8:     branches: [master]
9:   schedule:
10:     - cron: '0 17 * * 4'
11: 
12: permissions:
13:   contents: read
14: 
15: jobs:
16:   analyse:
17:     permissions:
18:       security-events: write
19:     name: Analyse
20:     runs-on: ubuntu-latest
21: 
22:     steps:
23:     - name: Checkout repository
24:       uses: actions/checkout@v3
25:       with:
26:         # We must fetch at least the immediate parents so that if this is
27:         # a pull request then we can checkout the head.
28:         fetch-depth: 2
29: 
30:     # If this run was triggered by a pull request event, then checkout
31:     # the head of the pull request instead of the merge commit.
32:     - run: git checkout HEAD^2
33:       if: ${{ github.event_name == 'pull_request' }}
34: 
35:     # Initializes the CodeQL tools for scanning.
36:     - name: Initialize CodeQL
37:       uses: github/codeql-action/init@v2
38:       with:
39:         queries: +security-extended
40:       # Override language selection by uncommenting this and choosing your languages
41:       # with:
42:       #   languages: go, javascript, csharp, python, cpp, java
43: 
44:     # Autobuild attempts to build any compiled languages  (C/C++, C#, or Java).
45:     # If this step fails, then you should remove it and run the build manually (see below)
46:     - name: Autobuild
47:       uses: github/codeql-action/autobuild@v2
48: 
49:     # ℹ️ Command-line programs to run using the OS shell.
50:     # 📚 https://git.io/JvXDl
51: 
52:     # ✏️ If the Autobuild fails above, remove it and uncomment the following three lines
53:     #    and modify them (or add more) to build your code if your project
54:     #    uses a compiled language
55: 
56:     #- run: |
57:     #   make bootstrap
58:     #   make release
59: 
60:     - name: Perform CodeQL Analysis
61:       uses: github/codeql-action/analyze@v2
62: 
```

- **A** (claude-opus-5-5): `real_defect` — rule: R1 Presence — The excerpt literally shows actions referenced by mutable version tags (actions/checkout@v3, github/codeql-action/*@v2) in a shipped, hand-written workflow file, with no SHA pinning or other neutralising construct visible. The .github/workflows path is CI configuration, not a test, fixture, example, docs or generated file, so R3 does not apply.
- **B** (gpt-6.1-sol): `real_defect` — rule: R1 Presence — The workflow references actions/checkout@v3 and github/codeql-action components @v2 instead of commit SHAs. These active action references have no visible pin or neutralising guard.

**Proposed label: `real_defect`**

Your decision: [ ] confirm  [ ] change to ______  note: ____
