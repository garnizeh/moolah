# Product Design Document: Moolah – Personal Finance Evolution

## 1. Executive Vision and Strategic Objectives
The transition of personal financial management from fragmented, manual spreadsheets to a centralized, API-driven architecture represents a strategic shift from retrospective logging to proactive wealth engineering. The legacy system, documented across "contas pagar" and "aposentadoria 2040," is hindered by manual entry risks, data drift, and "pendente" (pending) variances that obscure real-time net worth.
"Moolah" adopts a "Simple to Sophisticated" philosophy. The immediate objective is to translate legacy spreadsheet logic into a robust Golang-based API environment. This move minimizes human error—specifically targeting the systemic discrepancies in balance calculations—and establishes a scalable domain model. By migrating data into a structured system, we enable high-integrity tracking of expenditures and investments, providing the analytical foundation required to manage the journey from 2023 historicals to a 2040 retirement horizon.

## 2. Functional Domain: Accounts Payable & Expense Management
Categorized expenditure tracking is the primary requirement for household liquidity management. The system must move beyond simple ledger entries to capture the metadata required for deep cash-flow analysis.

### 2.1 Expense & Metadata Entity Definitions
To handle the complexity of the source data—where payment details are often concatenated—the system will implement a parsing layer to separate transaction intent from execution details.

| Field | Type | Mapping / Logic |
| :--- | :--- | :--- |
| Due Date | Date (DD/MM/YYYY) | Source "dia" column; used for aging reports. |
| Expected Value | Currency (R$) | Source "valor"; the contractually owed amount. |
| Amount Paid | Currency (R$) | Source "pago"; the actual cash outflow. |
| Status | Enum | Calculated: Paid, Partially Paid, or Overdue. |
| Entity Tag | String | Multi-entity support (e.g., Antônio, Sarah, Isabella). |
| TransactionMetadata | Object | Result of parsing the "pagamento" string. |

Parsing Logic for TransactionMetadata: The system must ingest strings like "06/01/23 pix mercadopago" and extract:
* Transaction Date: 06/01/2023
* Method: Pix
* Source Institution: MercadoPago

### 2.2 Financing & Installment Entity
The system must explicitly model long-term obligations (e.g., Volkswagen 11/36, Pemi 41/120) rather than treating them as static monthly costs.
* Total Installments: Integer (e.g., 36 or 120).
* Current Installment: Integer (auto-incrementing per month).
* Remaining Principal: Calculated based on Total - Current.

### 2.3 Categorization Hierarchy
The categorization engine must support sub-types to reflect complex household obligations found in the source:
* Domestic Help:
  * Salário: (Cris/Adriana)
  * Vale Transporte: (Specific tracking for transportation vouchers)
  * Impostos/eSocial: (Cris - imposto/13º taxes)
* Housing/Utilities: Aluguel, Sabesp, Energisa.
* Education: Antonio Jose - escola, Sarah - MBA EAD.
* Entity-Specific: Mesada (Luis Henrique, Isabella), Pensão Alimentícia (Split-entity logic).

## 3. Functional Domain: Investment Portfolio Tracking
The Investment Module shifts the focus from manual value updates to an automated "Single Source of Truth" for asset allocation.

### 3.1 Multi-Entity Asset Management
The architect must support separate portfolio tracking for Antônio (e.g., Antônio IPCA+ 2040) and the primary retirement account. This allows for specific goal-based reporting for dependents.

### 3.2 Suggestion Engine & Rebalancing
The API will calculate the variance between % Atual (Actual) and % Alvo (Target).
* Initial Position Logic: For assets like QBTC11, which shows a 0.00% actual weight against a 5% target, the engine must flag an "INITIAL_ENTRY" status to trigger the first purchase.
* Contribution Priority: Based on the source, where Tesouro IPCA 2040+ is over-weighted (91.69% vs 50% target), the engine must prioritize inflows into WRLD11 (8.31% vs 45% target) and QBTC11.

### 3.3 Transaction Logging & Cost Basis
The Golang backend must implement a Weighted Average Cost (WAC) formula for all assets.
* Data Capture: Capture Unit Price, Quantity, and Broker (Itaú, Inter, or Bradesco for IPVA obligations).
* Calculation: For the Tesouro IPCA 2040+ series (purchases on 15/01, 16/01, 19/01, etc.), the system will maintain the average price based on the cumulative total paid divided by total quantity held.

## 4. Analytical Layer: Trends and Forecasting
Strategic decision-making requires the system to identify shifts in household cost-of-living and project future stability.

### 4.1 "Pendente" Validation Algorithm
To ensure data integrity, the system will enforce an automated reconciliation check.
* Algorithm: Variance = abs(Sum(Expected_Value) - Sum(Amount_Paid))
* Enforcement: The system must flag any monthly variance (e.g., the R12.52 drift in 02/24 or R23.09 in 03/24) for manual reconciliation to prevent long-term accounting leakage.

### 4.2 Utility Variance & Inflation Tracking
The system will monitor variable utilities to identify cost spikes.
* Sabesp (Water): Tracking 02/23 (R133.10) through 02/26 (R375.00).
* Energisa (Power): Tracking 02/23 (R326.00) through 02/26 (R610.00). The forecasting engine will use a 12-month rolling average to project these costs into future budget months.

### 4.3 Family Evolution & Split Logic
The roadmap must prioritize evolving obligation structures. As seen in 04/25, the Pensão Alimentícia shifts from a single payment (Jaqueline) to a split-entity obligation (Jaqueline and Isabella). The system must support "Splitting Obligations" where one category evolves into two distinct legal/financial entities.

## 5. System Architecture & Product Roadmap

### 5.1 Technical Foundations (Phase 1: MVP)
* Golang REST API: Leveraging Go's concurrency for the Projection Service, using worker pools to run Monte Carlo simulations or inflation scenarios for the 2040 retirement target.
* Schema: PostgreSQL implementation of the parsed Expense, Financing, and Investment entities.
* CRUD Operations: Specialized endpoints for monthly expenditure entry with Pendente auto-calculation.

### 5.2 Expansion (Phase 2: Automated Intelligence)
* Accounts Receivable Module: Integration of salary and rent inflows to calculate net savings rate.
* Net Worth Visualizer: Historical tracking of asset growth (e.g., R195,203.77 in 02/26 to R258,740.00 in 03/26).
* Automated Forecasting: Worker-driven population of future months based on financing maturation dates and recurring utility trends.

## Conclusion
"Moolah" will replace the fragility of the legacy spreadsheets with a high-performance FinTech architecture. By enforcing strict data modeling for financing, multi-entity tracking for dependents like Antônio, and automated rebalancing logic, the application transforms raw historical data into a strategic asset for long-term wealth preservation.