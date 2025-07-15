import type { Route } from "./+types/screener";
import { useState, useEffect } from "react";
import {
  Form,
  Link,
  useSearchParams,
  useLoaderData,
  useNavigation,
} from "react-router";

// Meta function for SEO
export function meta({}: Route.MetaArgs) {
  return [
    { title: "Stock Screener - FinSights AI" },
    {
      name: "description",
      content:
        "Advanced stock screening tool with value investing metrics, dividend analysis, and technical indicators. Find undervalued stocks with custom filters.",
    },
    {
      name: "keywords",
      content:
        "stock screener, value investing, dividend stocks, P/E ratio, ROE, margin of safety, investment analysis",
    },
    { property: "og:title", content: "Stock Screener - FinSights AI" },
    {
      property: "og:description",
      content:
        "Find stocks that match your investment criteria using advanced filtering and valuation metrics.",
    },
    { property: "og:type", content: "website" },
  ];
}

// Types for our screener data
interface ScreenerResult {
  ticker: string;
  pe_ratio: number;
  roe: number;
  close: number;
  sma50: number;
  sma200: number;
  earnings_outlook: string;
  dividend_yield: number;
  dividend_growth_5y: number;
  intrinsic_value: number;
  margin_of_safety: number;
}

interface ScreenerResponse {
  data: ScreenerResult[];
  page: number;
  limit: number;
  total_count: number;
  has_more: boolean;
}

interface LoaderData {
  results: ScreenerResponse;
  filters: string;
  sort: string;
  page: number;
  limit: number;
}

// Loader function for server-side data fetching
export async function loader({
  request,
}: Route.LoaderArgs): Promise<LoaderData> {
  const url = new URL(request.url);
  const filters = url.searchParams.get("filters") || "";
  const sort = url.searchParams.get("sort") || "pe_ratio.asc";
  const page = parseInt(url.searchParams.get("page") || "1", 10);
  const limit = parseInt(url.searchParams.get("limit") || "25", 10);

  try {
    // Build API URL - adjust the backend URL as needed
    const backendPort = process.env.BACKEND_PORT || "8080";
    const apiUrl = new URL(`http://localhost:${backendPort}/api/screener`);
    if (filters) apiUrl.searchParams.set("filters", filters);
    apiUrl.searchParams.set("sort", sort);
    apiUrl.searchParams.set("page", page.toString());
    apiUrl.searchParams.set("limit", limit.toString());

    const response = await fetch(apiUrl.toString());

    if (!response.ok) {
      throw new Error(`API Error: ${response.status}`);
    }

    const results: ScreenerResponse = await response.json();

    return {
      results,
      filters,
      sort,
      page,
      limit,
    };
  } catch (error) {
    console.error("Failed to fetch screener data:", error);
    // Return empty results on error
    return {
      results: {
        data: [],
        page: 1,
        limit: 25,
        total_count: 0,
        has_more: false,
      },
      filters,
      sort,
      page,
      limit,
    };
  }
}

// Preset filter configurations
const PRESET_FILTERS = {
  value_stocks: {
    name: "Value Stocks",
    description: "Low P/E ratio with high ROE",
    filters: '[["pe_ratio","<",15],["roe",">",0.15]]',
    color: "bg-blue-100 text-blue-800 border-blue-200",
  },
  dividend_stocks: {
    name: "Dividend Stocks",
    description: "High dividend yield and growth",
    filters: '[["dividend_yield",">",0.03],["dividend_growth_5y",">",0.05]]',
    color: "bg-green-100 text-green-800 border-green-200",
  },
  undervalued_stocks: {
    name: "Undervalued Stocks",
    description: "High margin of safety",
    filters: '[["margin_of_safety",">",0.20]]',
    color: "bg-purple-100 text-purple-800 border-purple-200",
  },
  growth_stocks: {
    name: "Growth Stocks",
    description: "High ROE with positive outlook",
    filters: '[["roe",">",0.20],["earnings_outlook","=","positive"]]',
    color: "bg-orange-100 text-orange-800 border-orange-200",
  },
  bargain_stocks: {
    name: "Bargain Stocks",
    description: "Low P/E, below 200-day MA",
    filters: '[["pe_ratio","<",10],["price_vs_sma200","<",1.0]]',
    color: "bg-red-100 text-red-800 border-red-200",
  },
};

const SORT_OPTIONS = [
  { value: "pe_ratio.asc", label: "P/E Ratio (Low to High)" },
  { value: "pe_ratio.desc", label: "P/E Ratio (High to Low)" },
  { value: "roe.desc", label: "ROE (High to Low)" },
  { value: "roe.asc", label: "ROE (Low to High)" },
  { value: "dividend_yield.desc", label: "Dividend Yield (High to Low)" },
  { value: "margin_of_safety.desc", label: "Margin of Safety (High to Low)" },
  { value: "close.desc", label: "Price (High to Low)" },
  { value: "close.asc", label: "Price (Low to High)" },
  { value: "ticker.asc", label: "Ticker (A-Z)" },
];

export default function Screener() {
  const loaderData = useLoaderData<typeof loader>();
  const navigation = useNavigation();
  const [searchParams] = useSearchParams();

  // Local state for filter builder
  const [filterConditions, setFilterConditions] = useState<
    Array<{
      field: string;
      operator: string;
      value: string;
    }>
  >([]);
  const [showAdvancedFilters, setShowAdvancedFilters] = useState(false);

  const isLoading = navigation.state === "loading";

  // Parse existing filters on mount
  useEffect(() => {
    const filtersParam = searchParams.get("filters");
    if (filtersParam) {
      try {
        const parsed = JSON.parse(filtersParam);
        if (Array.isArray(parsed)) {
          setFilterConditions(
            parsed.map(([field, operator, value]) => ({
              field,
              operator,
              value: value.toString(),
            }))
          );
        }
      } catch (e) {
        console.error("Failed to parse filters:", e);
      }
    }
  }, [searchParams]);

  const addFilterCondition = () => {
    setFilterConditions([
      ...filterConditions,
      { field: "pe_ratio", operator: "<", value: "20" },
    ]);
  };

  const removeFilterCondition = (index: number) => {
    setFilterConditions(filterConditions.filter((_, i) => i !== index));
  };

  const updateFilterCondition = (
    index: number,
    field: keyof (typeof filterConditions)[0],
    value: string
  ) => {
    const updated = [...filterConditions];
    updated[index] = { ...updated[index], [field]: value };
    setFilterConditions(updated);
  };

  const buildFiltersJSON = () => {
    if (filterConditions.length === 0) return "";

    const conditions = filterConditions.map(({ field, operator, value }) => {
      // Convert value to appropriate type
      const numValue = parseFloat(value);
      const finalValue = isNaN(numValue) ? value : numValue;
      return [field, operator, finalValue];
    });

    return JSON.stringify(conditions);
  };

  const formatCurrency = (value: number) => {
    return new Intl.NumberFormat("en-US", {
      style: "currency",
      currency: "USD",
    }).format(value);
  };

  const formatPercentage = (value: number) => {
    return new Intl.NumberFormat("en-US", {
      style: "percent",
      minimumFractionDigits: 2,
    }).format(value);
  };

  const getOutlookBadgeColor = (outlook: string) => {
    switch (outlook.toLowerCase()) {
      case "positive":
        return "bg-green-100 text-green-800";
      case "negative":
        return "bg-red-100 text-red-800";
      case "neutral":
        return "bg-gray-100 text-gray-800";
      case "stable":
        return "bg-blue-100 text-blue-800";
      default:
        return "bg-gray-100 text-gray-800";
    }
  };

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Navigation */}
      <nav className="bg-white shadow-sm border-b">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center py-4">
            <div className="flex items-center space-x-4">
              <Link
                to="/"
                className="flex items-center text-blue-600 hover:text-blue-700 transition-colors"
              >
                <svg
                  className="w-5 h-5 mr-2"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"
                  />
                </svg>
                Home
              </Link>
              <span className="text-gray-300">|</span>
              <h1 className="text-2xl font-bold text-gray-900">
                Stock Screener
              </h1>
            </div>
          </div>
        </div>
      </nav>

      {/* Header */}
      <div className="bg-white shadow-sm border-b">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="py-6">
            <p className="text-gray-600">
              Find stocks that match your investment criteria
            </p>
          </div>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="grid grid-cols-1 lg:grid-cols-4 gap-8">
          {/* Filters Sidebar */}
          <div className="lg:col-span-1">
            <div className="bg-white rounded-lg shadow-sm border p-6 sticky top-8">
              <h2 className="text-lg font-semibold text-gray-900 mb-4">
                Filters
              </h2>

              {/* Preset Filters */}
              <div className="mb-6">
                <h3 className="text-sm font-medium text-gray-700 mb-3">
                  Quick Filters
                </h3>
                <div className="space-y-2">
                  {Object.entries(PRESET_FILTERS).map(([key, preset]) => (
                    <Form key={key} method="get" className="w-full">
                      <input
                        type="hidden"
                        name="filters"
                        value={preset.filters}
                      />
                      <input type="hidden" name="page" value="1" />
                      <button
                        type="submit"
                        className={`w-full text-left p-3 rounded-lg border transition-colors hover:shadow-sm ${preset.color}`}
                      >
                        <div className="font-medium text-sm">{preset.name}</div>
                        <div className="text-xs opacity-75 mt-1">
                          {preset.description}
                        </div>
                      </button>
                    </Form>
                  ))}
                </div>
              </div>

              {/* Advanced Filters */}
              <div className="mb-6">
                <button
                  type="button"
                  onClick={() => setShowAdvancedFilters(!showAdvancedFilters)}
                  className="flex items-center justify-between w-full text-sm font-medium text-gray-700 mb-3"
                >
                  <span>Advanced Filters</span>
                  <svg
                    className={`w-4 h-4 transition-transform ${
                      showAdvancedFilters ? "transform rotate-180" : ""
                    }`}
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M19 9l-7 7-7-7"
                    />
                  </svg>
                </button>

                {showAdvancedFilters && (
                  <div className="space-y-3">
                    {filterConditions.map((condition, index) => (
                      <div
                        key={index}
                        className="border rounded-lg p-3 bg-gray-50 text-gray-700"
                      >
                        <div className="grid grid-cols-12 gap-2 mb-2">
                          <select
                            value={condition.field}
                            onChange={(e) =>
                              updateFilterCondition(
                                index,
                                "field",
                                e.target.value
                              )
                            }
                            className="col-span-6 text-xs px-2 py-1 border rounded focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                          >
                            <option value="pe_ratio">P/E Ratio</option>
                            <option value="roe">ROE</option>
                            <option value="dividend_yield">
                              Dividend Yield
                            </option>
                            <option value="margin_of_safety">
                              Margin of Safety
                            </option>
                            <option value="close">Price</option>
                            <option value="earnings_outlook">Outlook</option>
                          </select>
                          <select
                            value={condition.operator}
                            onChange={(e) =>
                              updateFilterCondition(
                                index,
                                "operator",
                                e.target.value
                              )
                            }
                            className="col-span-5 text-xs px-2 py-1 border rounded focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                          >
                            <option value="<">Less than</option>
                            <option value=">">Greater than</option>
                            <option value="=">=Equal to</option>
                            <option value="<=">≤</option>
                            <option value=">=">≥</option>
                          </select>
                          <button
                            type="button"
                            onClick={() => removeFilterCondition(index)}
                            className="col-span-1 text-red-500 hover:text-red-700"
                          >
                            ×
                          </button>
                        </div>
                        <input
                          type="text"
                          value={condition.value}
                          onChange={(e) =>
                            updateFilterCondition(
                              index,
                              "value",
                              e.target.value
                            )
                          }
                          placeholder="Value"
                          className="w-full text-xs px-2 py-1 border rounded focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                        />
                      </div>
                    ))}

                    <button
                      type="button"
                      onClick={addFilterCondition}
                      className="w-full px-3 py-2 text-sm text-blue-600 border border-blue-200 rounded-lg hover:bg-blue-50 transition-colors"
                    >
                      + Add Filter
                    </button>

                    <Form method="get">
                      <input
                        type="hidden"
                        name="filters"
                        value={buildFiltersJSON()}
                      />
                      <input type="hidden" name="page" value="1" />
                      <button
                        type="submit"
                        className="w-full px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-medium"
                      >
                        Apply Filters
                      </button>
                    </Form>
                  </div>
                )}
              </div>

              {/* Clear Filters */}
              <Form method="get">
                <button
                  type="submit"
                  className="w-full px-4 py-2 text-sm text-gray-600 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
                >
                  Clear All Filters
                </button>
              </Form>
            </div>
          </div>

          {/* Results */}
          <div className="lg:col-span-3">
            {/* Controls */}
            <div className="bg-white rounded-lg shadow-sm border p-4 mb-6">
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-4">
                  <Form method="get" className="flex items-center space-x-2">
                    <input
                      type="hidden"
                      name="filters"
                      value={loaderData.filters}
                    />
                    <input type="hidden" name="page" value="1" />
                    <label
                      htmlFor="sort"
                      className="text-sm font-medium text-gray-700"
                    >
                      Sort by:
                    </label>
                    <select
                      id="sort"
                      name="sort"
                      value={loaderData.sort}
                      onChange={(e) => e.target.form?.submit()}
                      className="text-sm border border-gray-300 rounded-lg px-3 py-1 focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                    >
                      {SORT_OPTIONS.map((option) => (
                        <option key={option.value} value={option.value}>
                          {option.label}
                        </option>
                      ))}
                    </select>
                  </Form>

                  <Form method="get" className="flex items-center space-x-2">
                    <input
                      type="hidden"
                      name="filters"
                      value={loaderData.filters}
                    />
                    <input type="hidden" name="sort" value={loaderData.sort} />
                    <input type="hidden" name="page" value="1" />
                    <label
                      htmlFor="limit"
                      className="text-sm font-medium text-gray-700"
                    >
                      Show:
                    </label>
                    <select
                      id="limit"
                      name="limit"
                      value={loaderData.limit}
                      onChange={(e) => e.target.form?.submit()}
                      className="text-sm border border-gray-300 rounded-lg px-3 py-1 focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                    >
                      <option value="10">10</option>
                      <option value="25">25</option>
                      <option value="50">50</option>
                      <option value="100">100</option>
                    </select>
                  </Form>
                </div>

                <div className="text-sm text-gray-600">
                  {loaderData.results.total_count} results
                  {isLoading && (
                    <span className="ml-2 inline-flex items-center">
                      <svg
                        className="animate-spin h-4 w-4 text-blue-500"
                        fill="none"
                        viewBox="0 0 24 24"
                      >
                        <circle
                          className="opacity-25"
                          cx="12"
                          cy="12"
                          r="10"
                          stroke="currentColor"
                          strokeWidth="4"
                        ></circle>
                        <path
                          className="opacity-75"
                          fill="currentColor"
                          d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                        ></path>
                      </svg>
                    </span>
                  )}
                </div>
              </div>
            </div>

            {/* Results Table */}
            <div className="bg-white rounded-lg shadow-sm border overflow-hidden">
              {loaderData.results.data.length === 0 ? (
                <div className="text-center py-12">
                  <svg
                    className="mx-auto h-12 w-12 text-gray-400"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
                    />
                  </svg>
                  <h3 className="mt-2 text-sm font-medium text-gray-900">
                    No stocks found
                  </h3>
                  <p className="mt-1 text-sm text-gray-500">
                    Try adjusting your filters to find more results.
                  </p>
                </div>
              ) : (
                <div className="overflow-x-auto">
                  <table className="min-w-full divide-y divide-gray-200">
                    <thead className="bg-gray-50">
                      <tr>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          Ticker
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          Price
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          P/E
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          ROE
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          Div Yield
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          Safety Margin
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          Outlook
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          Technical
                        </th>
                      </tr>
                    </thead>
                    <tbody className="bg-white divide-y divide-gray-200">
                      {loaderData.results.data.map((stock) => (
                        <tr key={stock.ticker} className="hover:bg-gray-50">
                          <td className="px-6 py-4 whitespace-nowrap">
                            <div className="font-medium text-gray-900">
                              {stock.ticker}
                            </div>
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                            {formatCurrency(stock.close)}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                            {stock.pe_ratio.toFixed(1)}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                            {formatPercentage(stock.roe)}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                            {formatPercentage(stock.dividend_yield)}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap">
                            <span
                              className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                                stock.margin_of_safety > 0.2
                                  ? "bg-green-100 text-green-800"
                                  : stock.margin_of_safety > 0.1
                                  ? "bg-yellow-100 text-yellow-800"
                                  : "bg-red-100 text-red-800"
                              }`}
                            >
                              {formatPercentage(stock.margin_of_safety)}
                            </span>
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap">
                            <span
                              className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${getOutlookBadgeColor(
                                stock.earnings_outlook
                              )}`}
                            >
                              {stock.earnings_outlook}
                            </span>
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                            <div className="flex space-x-1">
                              {stock.close > stock.sma50 && (
                                <span className="inline-flex px-1 py-0.5 text-xs bg-blue-100 text-blue-800 rounded">
                                  SMA50↗
                                </span>
                              )}
                              {stock.close > stock.sma200 && (
                                <span className="inline-flex px-1 py-0.5 text-xs bg-green-100 text-green-800 rounded">
                                  SMA200↗
                                </span>
                              )}
                            </div>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>

            {/* Pagination */}
            {loaderData.results.data.length > 0 && (
              <div className="mt-6 flex items-center justify-between">
                <div className="text-sm text-gray-700">
                  Page {loaderData.page} of{" "}
                  {Math.ceil(loaderData.results.total_count / loaderData.limit)}
                </div>

                <div className="flex space-x-2">
                  {loaderData.page > 1 && (
                    <Form method="get">
                      <input
                        type="hidden"
                        name="filters"
                        value={loaderData.filters}
                      />
                      <input
                        type="hidden"
                        name="sort"
                        value={loaderData.sort}
                      />
                      <input
                        type="hidden"
                        name="limit"
                        value={loaderData.limit}
                      />
                      <input
                        type="hidden"
                        name="page"
                        value={loaderData.page - 1}
                      />
                      <button
                        type="submit"
                        className="px-3 py-2 text-sm bg-white border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
                      >
                        Previous
                      </button>
                    </Form>
                  )}

                  {loaderData.results.has_more && (
                    <Form method="get">
                      <input
                        type="hidden"
                        name="filters"
                        value={loaderData.filters}
                      />
                      <input
                        type="hidden"
                        name="sort"
                        value={loaderData.sort}
                      />
                      <input
                        type="hidden"
                        name="limit"
                        value={loaderData.limit}
                      />
                      <input
                        type="hidden"
                        name="page"
                        value={loaderData.page + 1}
                      />
                      <button
                        type="submit"
                        className="px-3 py-2 text-sm bg-white border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
                      >
                        Next
                      </button>
                    </Form>
                  )}
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
