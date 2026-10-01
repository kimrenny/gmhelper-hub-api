# GMHelper Solution Contracts Specification: Math & Geometry

## 1. Overview & Architecture Boundary

This specification defines the authoritative JSON contracts for AI-generated Math and Geometry solutions in the GMHelper ecosystem.

The system boundary and lifecycle is:

```text
gmhelper-web (Client)
    │
    │  1. REST Submission (POST /api/v1/tasks/{type})
    ▼
gmhelper-api (.NET Monolith)
    │
    │  2. gRPC SolveProblemRequest { TaskId, ProblemType, Payload, UserId }
    ▼
gmhelper-hub-api (AI Solution Hub)
    │
    │  3. Solver constructs prompt from Task Payload
    │  4. Gemini generates structured JSON
    │  5. Hub validates & normalizes AI output (Untrusted -> Verified)
    │  6. gRPC SolveProblemResponse { TaskId, Status, Result (JSON), Success }
    ▼
gmhelper-api (Storage & Adaptation)
    │
    │  7. Adapts & persists task + solution to storage
    │  8. Serves solution via GET /api/v1/tasks/{type}/{id}
    ▼
gmhelper-web (Rendering)
```

---

## 2. Field Classification Taxonomy

To prevent ambiguity, every field in both Math and Geometry schemas is classified under one of four roles:

1. **Canonical**: The authoritative, semantic source of truth produced and verified by the system.
2. **Derived**: Computed directly by the solver/server from canonical facts or intermediate logic.
3. **Compatibility**: Generated specifically to support existing legacy frontend renderers (e.g. single-block KaTeX strings) without breaking existing client components.
4. **Optional**: Supplementary context that may be omitted if not applicable to the problem type.

---

## 3. Math Solution Contract

### 3.1 Existing Math Flow & Contract
* **Submission**: `POST /api/v1/tasks/math` with payload `{"data": "<latex_string>"}`.
* **Storage** (`wwwroot/Tasks/Math/{id}.json`): Stored as `{"data": "<latex_string>"}`.
* **Current Frontend Expectation**: `MathCanvasSolutionService` expects `res.data.data` as a single LaTeX string, passed directly to `LatexRendererService.renderLatex()` for KaTeX rendering inside `\begin{aligned} ... \end{aligned}`.

### 3.2 Proposed Math Solution Schema (`SolveProblemResponse.result`)

```json
{
  "problemType": "math",
  "status": "completed",
  "problem": "2x + 5 = 15",
  "latexProblem": "2x + 5 = 15",
  "steps": [
    {
      "stepNumber": 1,
      "title": "Subtract 5 from both sides",
      "explanation": "Subtract 5 from both sides of the equation to isolate the linear term.",
      "latexFormula": "2x + 5 - 5 = 15 - 5 \\implies 2x = 10"
    },
    {
      "stepNumber": 2,
      "title": "Divide by 2",
      "explanation": "Divide both sides by the coefficient of x to find the value of the variable.",
      "latexFormula": "\\frac{2x}{2} = \\frac{10}{2} \\implies x = 5"
    }
  ],
  "finalAnswer": "x = 5",
  "latexAnswer": "x = 5",
  "compositeLatex": "2x + 5 = 15 \\\\\n2x = 10 \\\\\nx = 5"
}
```

### 3.3 Math Field Taxonomy & Justification

| Field | Taxonomy | Type | Required | Description & Rationale |
| ----- | -------- | ---- | :------: | ----------------------- |
| `problemType` | **Canonical** | string | Yes | Constant `"math"`. Identifies domain solver. |
| `status` | **Canonical** | string | Yes | Constant `"completed"`. |
| `problem` | **Canonical** | string | Yes | Clean textual / unicode problem statement. Used for plain logging and search indexing. |
| `latexProblem` | **Canonical** | string | Yes | Original LaTeX expression. Serves as the visual problem statement in math typography. |
| `steps` | **Canonical** | array | Yes | Discrete sequence of reasoning steps. Enables future step-by-step interactive cards. |
| `steps[].stepNumber` | **Canonical** | integer | Yes | 1-based sequential step index ($1, 2, \dots, n$). |
| `steps[].title` | **Canonical** | string | Yes | Short summary of the algebraic/analytical transformation. |
| `steps[].explanation` | **Canonical** | string | Yes | Pedagogical explanation of why and how the step is performed. |
| `steps[].latexFormula` | **Canonical** | string | Yes | Valid KaTeX snippet representing the mathematical state of this specific step. |
| `finalAnswer` | **Canonical** | string | Yes | Plain-text / unicode final result. Necessary for accessibility, speech synthesis, and plain-text UI summaries. |
| `latexAnswer` | **Canonical** | string | Yes | High-precision LaTeX representation of the final answer (e.g. `x = \\frac{1}{2}`). Distinct from `finalAnswer` which lacks formatting commands. |
| `compositeLatex` | **Compatibility** | string | Yes | Multiline LaTeX string formatted with `\\` line breaks. **Required for zero-downtime compatibility** with existing `gmhelper-web` `LatexRendererService`. |

---

## 4. Geometry Solution Contract

### 4.1 Existing Geometry Flow & Structure
* **Canvas Input**: Serialized figure dictionary containing `tool`, `path`, `lines`, `angles`, `elements`, `points`.
* **Current Storage** (`wwwroot/Tasks/Geo/{id}.json`):
  ```json
  {
    "task": { /* Raw canvas figure dictionary */ },
    "given": "ABC – triangle\nAB = BC = 5\nAC = 6\n",
    "solution": {},
    "answer": "..."
  }
  ```

---

### 4.2 Separation of Input Facts vs. AI Deductions

A fundamental design requirement is that **Gemini output is untrusted**. Gemini cannot be permitted to hallucinate or alter facts supplied by the user.

The Geometry schema enforces strict partitioning between four concepts:
1. **`inputFacts`**: The immutable geometric ground truth drawn by the user and verified by the backend.
2. **`target`**: The specific unknown(s) requested by the user or problem statement.
3. **`derivedFacts`**: Geometric deductions, newly constructed auxiliary points/lines, calculated angle measures, and calculated side lengths discovered by the AI reasoning process.
4. **`steps`**: The logical, step-by-step proof or calculation path.

---

### 4.3 Mathematically Consistent Geometry Reference Example

#### Problem Specification
* **Figure**: Triangle $ABC$ with side lengths $AB = 5$, $BC = 5$, $AC = 6$.
* **Target**: Calculate the Area ($S$), Perimeter ($P$), and Height ($BH$) perpendicular to base $AC$.

#### Rigorous Mathematical Proof & Verification
1. **Classification**: Triangle $ABC$ is isosceles with base $AC$ because $AB = BC = 5 \neq AC = 6$.
2. **Auxiliary Construction**: Draw altitude $BH \perp AC$ where $H \in AC$.
3. **Base Bisection**: In an isosceles triangle, the altitude to the base is also the median. Therefore, $H$ is the midpoint of $AC$:
   $$AH = HC = \frac{AC}{2} = \frac{6}{2} = 3$$
4. **Pythagorean Theorem in $\triangle ABH$**:
   $$AB^2 = AH^2 + BH^2 \implies 5^2 = 3^2 + BH^2 \implies 25 = 9 + BH^2 \implies BH^2 = 16 \implies BH = 4$$
5. **Area Calculation**:
   $$S = \frac{1}{2} \cdot AC \cdot BH = \frac{1}{2} \cdot 6 \cdot 4 = 12$$
6. **Perimeter Calculation**:
   $$P = AB + BC + AC = 5 + 5 + 6 = 16$$
7. **Angle Verification**:
   $$\cos(\angle BAC) = \frac{AH}{AB} = \frac{3}{5} = 0.6 \implies \angle BAC = \angle BCA = \arccos(0.6) \approx 53.13^\circ$$
   $$\angle ABC = 180^\circ - 2 \cdot 53.13^\circ = 73.74^\circ$$
   *(All angles sum to $53.13^\circ + 53.13^\circ + 73.74^\circ = 180.00^\circ$)*.

---

### 4.4 Proposed Geometry Solution Schema (`SolveProblemResponse.result`)

```json
{
  "problemType": "geometry",
  "status": "completed",
  "problemStatement": "In isosceles triangle ABC with side lengths AB = 5, BC = 5, and base AC = 6, find the altitude BH to base AC, the area S, and the perimeter P.",
  "inputFacts": {
    "figures": [
      {
        "id": "triangle_1",
        "type": "triangle",
        "vertices": ["A", "B", "C"]
      }
    ],
    "lengths": {
      "AB": 5.0,
      "BC": 5.0,
      "AC": 6.0
    },
    "angles": {}
  },
  "target": {
    "descriptions": [
      "Altitude BH to base AC",
      "Area S of triangle ABC",
      "Perimeter P of triangle ABC"
    ],
    "variables": ["BH", "S", "P"]
  },
  "derivedFacts": {
    "auxiliaryConstructions": [
      {
        "type": "altitude",
        "label": "BH",
        "fromVertex": "B",
        "toSegment": "AC",
        "footPoint": "H"
      }
    ],
    "lengths": {
      "AH": 3.0,
      "HC": 3.0,
      "BH": 4.0
    },
    "angles": {
      "BAC": 53.13,
      "BCA": 53.13,
      "ABC": 73.74,
      "AHB": 90.0,
      "BHC": 90.0
    },
    "metrics": {
      "perimeter": 16.0,
      "area": 12.0
    }
  },
  "steps": [
    {
      "stepNumber": 1,
      "title": "Construct Altitude and Determine Segment Lengths",
      "explanation": "Construct altitude BH perpendicular to base AC. In isosceles triangle ABC with AB = BC, altitude BH bisects base AC. Thus, H is the midpoint of AC.",
      "latexFormula": "AH = HC = \\frac{AC}{2} = \\frac{6}{2} = 3"
    },
    {
      "stepNumber": 2,
      "title": "Apply Pythagorean Theorem in Right Triangle ABH",
      "explanation": "In right-angled triangle ABH, by the Pythagorean theorem, the square of hypotenuse AB equals the sum of the squares of legs AH and BH.",
      "latexFormula": "AB^2 = AH^2 + BH^2 \\implies 5^2 = 3^2 + BH^2 \\implies BH = \\sqrt{25 - 9} = 4"
    },
    {
      "stepNumber": 3,
      "title": "Calculate Area of Triangle ABC",
      "explanation": "The area S of a triangle is half the product of its base and corresponding altitude.",
      "latexFormula": "S = \\frac{1}{2} \\cdot AC \\cdot BH = \\frac{1}{2} \\cdot 6 \\cdot 4 = 12"
    },
    {
      "stepNumber": 4,
      "title": "Calculate Perimeter of Triangle ABC",
      "explanation": "The perimeter P is the sum of all three side lengths.",
      "latexFormula": "P = AB + BC + AC = 5 + 5 + 6 = 16"
    }
  ],
  "finalAnswer": "Altitude BH = 4, Area S = 12, Perimeter P = 16",
  "latexAnswer": "BH = 4, \\quad S = 12, \\quad P = 16"
}
```

---

### 4.5 Geometry Field Taxonomy & Justification

| Field | Taxonomy | Type | Required | Description & Rationale |
| ----- | -------- | ---- | :------: | ----------------------- |
| `problemType` | **Canonical** | string | Yes | Constant `"geometry"`. |
| `status` | **Canonical** | string | Yes | `"completed"` or `"unsolvable"`. |
| `problemStatement` | **Canonical** | string | Yes | Comprehensive natural language description of the problem and target. |
| `inputFacts` | **Canonical** | object | Yes | Verified input facts extracted from the user's canvas. Must match the input payload. |
| `inputFacts.figures` | **Canonical** | array | Yes | List of geometric figures drawn by the user. |
| `inputFacts.lengths` | **Canonical** | map | Yes | User-provided line lengths (e.g. `{"AB": 5.0}`). |
| `inputFacts.angles` | **Canonical** | map | Yes | User-provided angle measures in degrees. |
| `target` | **Canonical** | object | Yes | Structured declaration of target variables and natural-language goal. |
| `derivedFacts` | **Derived** | object | Yes | Strongly-typed container for AI deductions. Replaces the former untyped `calculatedProperties`. |
| `derivedFacts.auxiliaryConstructions` | **Derived** | array | No | Explicit record of auxiliary points/lines constructed by the proof (e.g. altitude foot $H$). |
| `derivedFacts.lengths` | **Derived** | map | No | Calculated line lengths discovered during the solution. |
| `derivedFacts.angles` | **Derived** | map | No | Calculated angle degrees discovered during the solution. |
| `derivedFacts.metrics` | **Derived** | map | Yes | Solved global properties (e.g. `area`, `perimeter`, `radius`). |
| `steps` | **Canonical** | array | Yes | Sequential proof and calculation steps. |
| `steps[].stepNumber` | **Canonical** | integer | Yes | 1-based sequential step index ($1, 2, \dots, n$). |
| `steps[].title` | **Canonical** | string | Yes | Geometric theorem or principle applied (e.g. `"Pythagorean Theorem"`). |
| `steps[].explanation` | **Canonical** | string | Yes | Rigorous geometric deduction and justification. |
| `steps[].latexFormula` | **Canonical** | string | Yes | KaTeX formula showing symbolic equation and arithmetic resolution. |
| `finalAnswer` | **Canonical** | string | Yes | Clean plain-text final statement. |
| `latexAnswer` | **Canonical** | string | Yes | Formatted LaTeX answer string. |

---

## 5. Justification: Typed `derivedFacts` vs. Untyped `calculatedProperties`

The previous draft proposed an open dictionary `calculatedProperties: { ... }`. This has been replaced with the strongly-typed `derivedFacts` structure for the following reasons:

1. **Deterministic Deserialization**: In Go, C#, and TypeScript, untyped maps (`map[string]interface{}`) require brittle runtime reflection and type-casting. Structured types (`AuxiliaryConstruction`, `Lengths`, `Angles`, `Metrics`) map directly to typed DTOs.
2. **Server-Side Verifiability**: A structured map `derivedFacts.lengths` allows the hub validator to perform mathematical consistency checks (e.g. verifying $AH + HC == AC$, $BH > 0$) without parsing arbitrary keys.
3. **Figure Graph Integrity**: The `auxiliaryConstructions` array tells the frontend explicitly how new points ($H$) and segments ($BH$) connect to existing vertices ($A, B, C$) without corrupting the original drawing.

---

## 6. Validation Requirements & Server Invariants

Gemini output is treated as untrusted data. Before `gmhelper-hub-api` accepts a solution, it must pass the following validation rules:

### 6.1 Common Invariants
1. **Valid JSON**: Payload must unmarshal into the target Go struct without error.
2. **Payload Size Limit**: Maximum response payload size is 64 KB.
3. **Step Sequentiality**: `steps` array must have length $\ge 1$ and step numbers must strictly increase from $1$ to $N$.
4. **Non-Empty Text**: `problemStatement`, `finalAnswer`, and `latexAnswer` must be non-empty after trimming whitespace.
5. **No Forbidden Tags**: Response must not contain markdown code fences (` ```json `), HTML tags (`<script>`), or system prompt echoes.

### 6.2 Math Validation Invariants
1. **KaTeX Brace Balance**: In `latexFormula`, `latexAnswer`, and `compositeLatex`, every `{` must have a matching `}`.
2. **KaTeX Command Whitelist**: Reject dangerous TeX control sequences (`\def`, `\let`, `\write`, `\input`, `\catcode`).
3. **Composite LaTeX Consistency**: `compositeLatex` must include line break delimiters (`\\`) if $N > 1$ steps exist.

### 6.3 Geometry Validation Invariants
1. **Input Fact Parity**: Every length/angle in `inputFacts` must match the ground truth provided in the original request. Gemini is not permitted to mutate user input facts.
2. **Vertex Reference Validity**: Every vertex referenced in `derivedFacts.auxiliaryConstructions` must be either an input vertex ($A, B, C$) or an explicitly defined `footPoint` / `auxiliaryPoint`.
3. **Positive Quantities**: All line lengths and area/perimeter metrics in `derivedFacts` must be finite positive numbers ($> 0$).
4. **Angle Ranges**: All angle values in `derivedFacts.angles` must fall strictly within $(0^\circ, 180^\circ)$.

---

## 7. Division of Responsibilities

```text
┌──────────────────────────────────────────────────────────────────────────────────┐
│                             GEMINI AI RESPONSIBILITIES                           │
├──────────────────────────────────────────────────────────────────────────────────┤
│ - Mathematical deduction, theorem selection, and equation solving.              │
│ - Natural language explanation generation.                                       │
│ - Formatting formulas in standard KaTeX notation.                                │
│ - Conforming to the requested JSON response schema.                              │
└──────────────────────────────────────────────────────────────────────────────────┘
                                       │
                                       ▼  (Raw JSON via Google GenAI SDK)
┌──────────────────────────────────────────────────────────────────────────────────┐
│                            HUB SERVER RESPONSIBILITIES                           │
├──────────────────────────────────────────────────────────────────────────────────┤
│ - JSON Unmarshaling into strongly-typed Go structures.                           │
│ - Strict invariant validation (positive lengths, angle limits, brace balance).  │
│ - Verifying input facts against original task.                                   │
│ - Generating backward-compatible fields (e.g. compositeLatex for math).          │
│ - Sanitizing output & mapping errors to safe gRPC status codes.                 │
└──────────────────────────────────────────────────────────────────────────────────┘
                                       │
                                       ▼  (gRPC SolveProblemResponse.result)
┌──────────────────────────────────────────────────────────────────────────────────┐
│                          API MONOLITH RESPONSIBILITIES                           │
├──────────────────────────────────────────────────────────────────────────────────┤
│ - Persisting task and solution records.                                          │
│ - Serving GET endpoints to web clients.                                          │
│ - Handling user authentication, rate limiting, and solution ratings.             │
└──────────────────────────────────────────────────────────────────────────────────┘
```

---

## 8. Persistence vs. AI Result Flow

The existing `gmhelper-api` storage format:

```json
{
  "task": { /* canvas input */ },
  "given": "...",
  "solution": {},
  "answer": "..."
}
```

remains fully compatible with the new AI solution result:

```text
1. User submits task -> gmhelper-api creates taskId.
2. gmhelper-api invokes gRPC SolveProblem(taskId, problemType, payload).
3. gmhelper-hub-api validates problem, invokes Gemini, verifies schema, returns SolveProblemResponse.
4. gmhelper-api receives validated AI JSON string in response.Result.
5. gmhelper-api stores:
   - For Math: { "data": aiResult.compositeLatex, "solution": aiResult }
   - For Geo:  { "task": inputTask, "given": aiResult.problemStatement, "solution": aiResult, "answer": aiResult.finalAnswer }
6. gmhelper-web fetches task via GET /api/v1/tasks/{type}/{id} and renders:
   - Legacy view: consumes data / task / answer directly without changes.
   - Enhanced view: consumes solution.steps[] and solution.derivedFacts for rich UI cards.
```

---

## 9. Open Questions for Future UI Stages

1. **Geometry Canvas Visual Highlighting**:
   * *Question*: When `derivedFacts.auxiliaryConstructions` contains an altitude (e.g. $BH$), should the frontend canvas dynamically draw a dashed line for $BH$ with foot label $H$, or should the visual canvas remain strictly as drawn by the user while the solution is displayed in an adjacent accordion card?
   * *Recommendation*: Keep canvas immutable in initial solver release and display constructed elements in the solution step cards. In a future UI iteration, add an optional "Show Construction on Canvas" toggle.
2. **Localization ($i18n$) for AI Explanations**:
   * *Question*: GMHelper supports multiple UI languages (English, Russian, etc.). Should the solver prompt request explanations in the user's preferred language (passed via `SolveProblemRequest`) or default to English?
   * *Recommendation*: Default to English in initial release, add optional `language` code to task request in subsequent iteration.
