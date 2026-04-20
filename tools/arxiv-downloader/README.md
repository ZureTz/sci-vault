# arxiv-downloader

A small Python CLI for fetching PDFs from [arXiv](https://arxiv.org) — either
by specific ID, by free-text query, or in bulk across one or more subject
categories. Uses the official [arXiv API](https://info.arxiv.org/help/api/index.html).

## Requirements

- Python **3.14+**
- [`uv`](https://docs.astral.sh/uv/) for environment management

## Setup

```bash
cd tools/arxiv-downloader
uv sync
```

`uv` will install the pinned Python version and dependencies into `.venv`.

## Usage

Run everything via `uv run` so the managed virtualenv is picked up:

```bash
uv run main.py [MODE] [OPTIONS]
```

There are three modes, selected by one of `--id`, `--query`, or `--fields`.

### 1. By arXiv ID

Download one or more specific papers. Accepts bare IDs, `arXiv:` prefixes, and
full abs/pdf URLs.

```bash
uv run main.py --id 1706.03762
uv run main.py --id 2401.12345 2402.54321 math.GT/0309136
uv run main.py --id https://arxiv.org/abs/1706.03762
```

### 2. By free-text query

Search the whole arXiv corpus and grab the top results (newest first).

```bash
uv run main.py --query "retrieval augmented generation" --max 10
uv run main.py --query "diffusion models for audio" --max 5
```

### 3. Batch by subject category (fields)

The main use case: fetch `N` papers spread across several arXiv categories.
Each field's PDFs go into their own subfolder under `--out`.

```bash
# 100 papers, evenly split across 5 fields
uv run main.py \
    --fields cs.AI cs.LG cs.CV stat.ML math.CO \
    --total 100

# Fixed count per field (25 each → 50 total)
uv run main.py \
    --fields cs.AI cs.CV \
    --per-field 25 \
    --out ./papers \
    --max-size-mb 50
```

`--fields` requires **exactly one** of `--total` or `--per-field`.

## Common options

| Flag             | Default      | Purpose                                                     |
| ---------------- | ------------ | ----------------------------------------------------------- |
| `--out DIR`      | `./papers`   | Output directory. Batch mode creates one subdir per field.  |
| `--max-size-mb N`| `50`         | Skip any PDF larger than `N` MB (checked via header + stream). |
| `--overwrite`    | off          | Re-download even if the target file already exists.         |
| `--max N`        | `10`         | Max results when using `--query`.                           |

Files are named `{arxiv_id} - {title}.pdf`, sanitized for filesystem safety.

## Rate limiting

The script is polite by default:

- **3 seconds** between arXiv API calls (metadata lookups).
- **1 second** between PDF downloads.

Don't lower these when running large batches — arXiv will throttle you.

## arXiv subject categories

Pass any of the codes below to `--fields`. The full, authoritative list lives
at <https://arxiv.org/category_taxonomy>. A curated selection of the most
commonly used codes:

### Computer Science (`cs.*`)

| Code       | Area                                          |
| ---------- | --------------------------------------------- |
| `cs.AI`    | Artificial Intelligence                       |
| `cs.AR`    | Hardware Architecture                         |
| `cs.CC`    | Computational Complexity                      |
| `cs.CE`    | Computational Engineering, Finance, Science   |
| `cs.CG`    | Computational Geometry                        |
| `cs.CL`    | Computation and Language (NLP)                |
| `cs.CR`    | Cryptography and Security                     |
| `cs.CV`    | Computer Vision and Pattern Recognition       |
| `cs.CY`    | Computers and Society                         |
| `cs.DB`    | Databases                                     |
| `cs.DC`    | Distributed, Parallel, and Cluster Computing  |
| `cs.DS`    | Data Structures and Algorithms                |
| `cs.GT`    | Computer Science and Game Theory              |
| `cs.HC`    | Human-Computer Interaction                    |
| `cs.IR`    | Information Retrieval                         |
| `cs.IT`    | Information Theory                            |
| `cs.LG`    | Machine Learning                              |
| `cs.LO`    | Logic in Computer Science                     |
| `cs.MA`    | Multiagent Systems                            |
| `cs.NE`    | Neural and Evolutionary Computing             |
| `cs.NI`    | Networking and Internet Architecture          |
| `cs.OS`    | Operating Systems                             |
| `cs.PL`    | Programming Languages                         |
| `cs.RO`    | Robotics                                      |
| `cs.SE`    | Software Engineering                          |
| `cs.SI`    | Social and Information Networks               |
| `cs.SY`    | Systems and Control                           |

### Mathematics (`math.*`)

| Code       | Area                       |
| ---------- | -------------------------- |
| `math.AG`  | Algebraic Geometry         |
| `math.AT`  | Algebraic Topology         |
| `math.CA`  | Classical Analysis and ODEs|
| `math.CO`  | Combinatorics              |
| `math.DG`  | Differential Geometry      |
| `math.DS`  | Dynamical Systems          |
| `math.FA`  | Functional Analysis        |
| `math.GT`  | Geometric Topology         |
| `math.LO`  | Logic                      |
| `math.NA`  | Numerical Analysis         |
| `math.NT`  | Number Theory              |
| `math.OC`  | Optimization and Control   |
| `math.PR`  | Probability                |
| `math.ST`  | Statistics Theory          |

### Statistics (`stat.*`)

| Code       | Area              |
| ---------- | ----------------- |
| `stat.AP`  | Applications      |
| `stat.CO`  | Computation       |
| `stat.ME`  | Methodology       |
| `stat.ML`  | Machine Learning  |
| `stat.TH`  | Theory            |

### Physics

| Code               | Area                                             |
| ------------------ | ------------------------------------------------ |
| `astro-ph.CO`      | Cosmology and Nongalactic Astrophysics           |
| `astro-ph.GA`      | Astrophysics of Galaxies                         |
| `astro-ph.HE`      | High Energy Astrophysical Phenomena              |
| `astro-ph.SR`      | Solar and Stellar Astrophysics                   |
| `cond-mat.mes-hall`| Mesoscale and Nanoscale Physics                  |
| `cond-mat.stat-mech`| Statistical Mechanics                           |
| `gr-qc`            | General Relativity and Quantum Cosmology         |
| `hep-ex`           | High Energy Physics — Experiment                 |
| `hep-ph`           | High Energy Physics — Phenomenology              |
| `hep-th`           | High Energy Physics — Theory                     |
| `math-ph`          | Mathematical Physics                             |
| `nucl-ex`          | Nuclear Experiment                               |
| `nucl-th`          | Nuclear Theory                                   |
| `physics.app-ph`   | Applied Physics                                  |
| `physics.bio-ph`   | Biological Physics                               |
| `physics.comp-ph`  | Computational Physics                            |
| `physics.data-an`  | Data Analysis, Statistics and Probability        |
| `physics.flu-dyn`  | Fluid Dynamics                                   |
| `physics.optics`   | Optics                                           |
| `quant-ph`         | Quantum Physics                                  |

### Quantitative Biology (`q-bio.*`)

| Code       | Area                              |
| ---------- | --------------------------------- |
| `q-bio.BM` | Biomolecules                      |
| `q-bio.GN` | Genomics                          |
| `q-bio.MN` | Molecular Networks                |
| `q-bio.NC` | Neurons and Cognition             |
| `q-bio.PE` | Populations and Evolution         |
| `q-bio.QM` | Quantitative Methods              |

### Quantitative Finance (`q-fin.*`)

| Code         | Area                       |
| ------------ | -------------------------- |
| `q-fin.CP`   | Computational Finance      |
| `q-fin.PM`   | Portfolio Management       |
| `q-fin.PR`   | Pricing of Securities      |
| `q-fin.RM`   | Risk Management            |
| `q-fin.ST`   | Statistical Finance        |
| `q-fin.TR`   | Trading and Market Microstructure |

### Electrical Engineering and Systems Science (`eess.*`)

| Code       | Area                          |
| ---------- | ----------------------------- |
| `eess.AS`  | Audio and Speech Processing   |
| `eess.IV`  | Image and Video Processing    |
| `eess.SP`  | Signal Processing             |
| `eess.SY`  | Systems and Control           |

### Economics (`econ.*`)

| Code        | Area                    |
| ----------- | ----------------------- |
| `econ.EM`   | Econometrics            |
| `econ.GN`   | General Economics       |
| `econ.TH`   | Theoretical Economics   |

## Troubleshooting

- **"No results"**: double-check the category code spelling. `cs.ai` is wrong;
  `cs.AI` is correct (suffix is case-sensitive).
- **HTTP 503 / rate limit**: wait a minute and re-run. arXiv occasionally
  throttles; the script resumes cleanly thanks to the existence check.
- **Oversize skips**: bump `--max-size-mb` if you legitimately want large files
  (e.g. thesis-length papers or data appendices).

## Exit codes

| Code | Meaning                                        |
| ---- | ---------------------------------------------- |
| `0`  | All requested papers downloaded or skipped-as-existing. |
| `1`  | No results found for the given query/IDs.      |
| `2`  | At least one download failed with an HTTP error. |
