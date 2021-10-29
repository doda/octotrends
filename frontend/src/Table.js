import React, { useEffect } from "react";
import {
  useTable,
  useFilters,
  useSortBy,
  usePagination,
  useGroupBy,
  useExpanded,
} from "react-table";
import {
  ChevronDoubleLeftIcon,
  ChevronLeftIcon,
  ChevronRightIcon,
  ChevronDoubleRightIcon,
} from "@heroicons/react/solid";
import { Button, PageButton } from "./shared/Button";
import { ButtonGroup, GroupButton } from "./shared/ButtonGroup";
import { humanNumber, replaceColonmoji } from "./shared/Utils";
import { SortIcon, SortUpIcon, SortDownIcon } from "./shared/Icons";
import { StarIcon, SquareFillIcon } from "@primer/octicons-react";
import Colors from "./colors.json";

// This is a custom filter UI for selecting
// a unique option from a list
export function SelectColumnFilter(props) {
  let {
    column: { filterValue, setFilter, preFilteredRows, id, render },
    state: { groupBy },
  } = props;

  // Calculate the options for filtering
  // using the preFilteredRows
  let options = React.useMemo(() => {
    const options = new Set();
    preFilteredRows.forEach((row) => {
      if (row.values[id] !== "") options.add(row.values[id]);
    });
    return [...options.values()];
  }, [id, preFilteredRows]);
  if (groupBy.length > 0) return null;

  // Render a multi-select box
  return (
    <label className="flex gap-x-2 items-baseline">
      <span className="text-gray-700">{render("Header")}: </span>
      <select
        className="rounded-md border-gray-300 shadow-sm focus:border-indigo-300 focus:ring focus:ring-indigo-200 focus:ring-opacity-50 w-48"
        name={id}
        id={id}
        value={filterValue}
        onChange={(e) => {
          setFilter(e.target.value || undefined);
        }}
      >
        <option value="">All</option>
        {options.map((option, i) => (
          <option key={i} value={option}>
            {option}
          </option>
        ))}
      </select>
    </label>
  );
}

export function filterGreaterThan(rows, id, filterValue) {
  return rows.filter((row) => {
    const rowValue = row.values[id];
    return rowValue >= filterValue;
  });
}

export function filterSmallerThan(rows, id, filterValue) {
  return rows.filter((row) => {
    const rowValue = row.values[id];
    return rowValue < filterValue;
  });
}

export function filterSizes(rows, id, filterValue) {
  let sizes = {
    XS: { min: 0, max: 1000 },
    S: { min: 1000, max: 5000 },
    M: { min: 5000, max: 20000 },
    L: { min: 20000, max: Number.MAX_SAFE_INTEGER },
  };
  let activeFilters = Object.keys(filterValue).filter(
    (key) => filterValue[key]
  );

  return rows.filter((row) => {
    const rowValue = row.values[id];
    return activeFilters.some(
      (size) => rowValue > sizes[size].min && rowValue <= sizes[size].max
    );
  });
}

export function SizeFilter(props) {
  let {
    column: { filterValue, setFilter },
    state: { groupBy },
  } = props;
  if (groupBy.length > 0) return null;
  let fValue = filterValue || {};
  return (
    <ButtonGroup className="justify-end sm:justify-center mt-3 pt-px sm:mt-px">
      <GroupButton
        left
        active={fValue.XS}
        onClick={(e) => setFilter({ ...fValue, ...{ XS: !fValue.XS } })}
      >
        &lt;1k
      </GroupButton>
      <GroupButton
        active={fValue.S}
        onClick={(e) => setFilter({ ...fValue, ...{ S: !fValue.S } })}
      >
        1k-5k
      </GroupButton>
      <GroupButton
        active={fValue.M}
        onClick={(e) => setFilter({ ...fValue, ...{ M: !fValue.M } })}
      >
        5k-20k
      </GroupButton>
      <GroupButton
        right
        active={fValue.L}
        onClick={(e) => setFilter({ ...fValue, ...{ L: !fValue.L } })}
      >
        &gt;20k
      </GroupButton>
    </ButtonGroup>
  );
}

export function NameCell({ value, row }) {
  if (value == null) return null;
  let [owner, name] = value.split("/");
  let description = replaceColonmoji((row.original || {}).Description);
  return (
    <div className="truncate" style={{ width: 300 }}>
      <a
        target="_blank"
        rel="noreferrer"
        title={description}
        href={"https://github.com/" + value}
      >
        <span className="text-sm text-blue-500">
          {owner}/<strong>{name}</strong>
        </span>
        <br />
        <span className="text-sm text-gray-500">{description}&nbsp;</span>
      </a>
    </div>
  );
}

export function LanguageCell({ value, setFilter, columns, state }) {
  if (value.length === 0) return null;
  let languageCol = columns.filter(function (entry) {
    return entry.id === "Language";
  })[0];

  let linkProps = {
    href: "#",
    className: "truncate hover:underline w-20",
    title: value,
    onClick: (e) => {
      // Untoggle group by language and show repos for this language
      if (state.groupBy.length > 0) {
        languageCol.getGroupByToggleProps().onClick(e);
      }
      setFilter("Language", value);
    },
  };

  return (
    <div className="text-sm truncate whitespace-nowrap w-32">
      <a {...linkProps}>
        {" "}
        <span style={{ color: (Colors[value] || {}).color }}>
          <SquareFillIcon />
        </span>
        <span className="text-gray-500 text-sm">{value}</span>
      </a>
    </div>
  );
}

export function StarCell({ value }) {
  return (
    <span className="text-gray-500 text-sm whitespace-nowrap">
      <StarIcon />
      {humanNumber(value)}
    </span>
  );
}

export function GrowthCell({ value }) {
  // value = growthCalc(value);
  return value === 0 ? null : (
    <span className="text-gray-500 text-sm">
      {value.added > 0 ? "+" : ""}
      {humanNumber(value.added)}
    </span>
  );
}

function Table({ columns, data }) {
  // Use the state and functions returned from useTable to build your UI
  const starsFilterDefault = { XS: true, S: true, M: true, L: true };
  const {
    getTableProps,
    getTableBodyProps,
    headerGroups,
    prepareRow,
    page, // Instead of using 'rows', we'll use page,
    // which has only the rows for the active page

    // The rest of these things are super handy, too ;)
    canPreviousPage,
    canNextPage,
    pageOptions,
    pageCount,
    gotoPage,
    nextPage,
    previousPage,
    setPageSize,
    setFilter,
    state,
  } = useTable(
    {
      columns,
      data,
      initialState: {
        pageSize: 10,
        sortBy: [
          {
            id: "Growth30",
            desc: true,
          },
        ],
        filters: [{ id: "Stars", value: starsFilterDefault }],
        hiddenColumns: [
          "data",
          "Added7",
          "Baseline7",
          "Added30",
          "Baseline30",
          "Added90",
          "Baseline90",
        ],
      },
    },
    useFilters,
    useGroupBy,
    useSortBy,
    useExpanded,
    usePagination
  );
  useEffect(() => {
    document.onkeydown = function (e) {
      switch (e.which) {
        case 37: // left
          previousPage();
          break;
        case 35: // end
          gotoPage(pageCount - 1);
          break;
        case 36: // home
          gotoPage(0);
          break;
        case 39: // right
          nextPage();
          break;

        default:
          return; // exit this handler for other keys
      }
      e.preventDefault(); // prevent the default action (scroll / move caret)
    };
  }, []);

  return (
    <>
      <div
        className="sm:flex sm:gap-x-2 h-36 sm:h-full"
        style={{
          minHeight: "2.65rem",
        }}
      >
        {headerGroups.map((headerGroup) =>
          headerGroup.headers.map((column) =>
            column.Filter ? (
              <div className="mt-2 sm:mt-0" key={column.id}>
                {column.render("Filter")}
              </div>
            ) : null
          )
        )}
      </div>
      {/* table */}
      <div className="mt-3 flex flex-col relative">
        <div className="-my-2 overflow-x-auto -mx-4 sm:-mx-6 lg:-mx-8">
          <div className="py-2 align-middle inline-block min-w-full sm:px-6 lg:px-6">
            <div className="shadow overflow-hidden border-b border-gray-200 sm:rounded-lg">
              <table
                {...getTableProps()}
                className="min-w-full divide-y divide-gray-200"
              >
                <thead className="bg-gray-50">
                  {headerGroups.map((headerGroup) => (
                    <tr {...headerGroup.getHeaderGroupProps()}>
                      {headerGroup.headers.map((column) => (
                        // Add the sorting props to control sorting. For this example
                        // we can add them into the header props
                        <th
                          scope="col"
                          className="group px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                          {...column.getHeaderProps(
                            column.getSortByToggleProps()
                          )}
                        >
                          <div className="flex items-center justify-between">
                            {column.canGroupBy ? (
                              <ButtonGroup
                                className="absolute right-0"
                                style={{ top: "-3.25rem" }}
                              >
                                <GroupButton
                                  left
                                  active={!column.isGrouped}
                                  onClick={(e) => {
                                    e.stopPropagation();
                                    e.nativeEvent.stopImmediatePropagation();

                                    column.getGroupByToggleProps().onClick(e);
                                    if (!column.isGrouped) {
                                      setFilter("Stars", starsFilterDefault);
                                      setFilter("Language", "");
                                    }
                                  }}
                                >
                                  Repos
                                </GroupButton>
                                <GroupButton
                                  right
                                  active={column.isGrouped}
                                  onClick={(e) => {
                                    e.stopPropagation();
                                    e.nativeEvent.stopImmediatePropagation();

                                    column.getGroupByToggleProps().onClick(e);
                                    if (!column.isGrouped) {
                                      setFilter("Stars", starsFilterDefault);
                                      setFilter("Language", "");
                                    }
                                  }}
                                >
                                  Languages
                                </GroupButton>
                              </ButtonGroup>
                            ) : null}

                            {column.render("Header")}
                            {/* Add a sort direction indicator */}
                            <span>
                              {column.isSorted ? (
                                column.isSortedDesc ? (
                                  <SortDownIcon className="w-4 h-4 text-gray-400" />
                                ) : (
                                  <SortUpIcon className="w-4 h-4 text-gray-400" />
                                )
                              ) : (
                                <SortIcon className="w-4 h-4 text-gray-400 opacity-0 group-hover:opacity-100" />
                              )}
                            </span>
                          </div>
                        </th>
                      ))}
                    </tr>
                  ))}
                </thead>
                <tbody
                  {...getTableBodyProps()}
                  className="bg-white divide-y divide-gray-200"
                >
                  {page.map((row, i) => {
                    // new
                    prepareRow(row);
                    return (
                      <tr {...row.getRowProps()}>
                        {row.cells.map((cell) => {
                          return (
                            <td
                              {...cell.getCellProps()}
                              className="px-6 py-3 whitespace-nowrap"
                              role="cell"
                            >
                              {cell.column.Cell.name === "defaultRenderer" ? (
                                <div className="text-sm text-gray-500">
                                  {cell.render("Cell")}
                                </div>
                              ) : (
                                cell.render("Cell")
                              )}
                            </td>
                          );
                        })}
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>
      {/* Pagination */}
      <div className="py-3 flex items-center justify-between">
        <div className="flex-1 flex justify-between sm:hidden">
          <Button onClick={() => previousPage()} disabled={!canPreviousPage}>
            Previous
          </Button>
          <Button onClick={() => nextPage()} disabled={!canNextPage}>
            Next
          </Button>
        </div>
        <div className="hidden sm:flex-1 sm:flex sm:items-center sm:justify-between">
          <div className="flex gap-x-2 items-baseline">
            <span className="text-sm text-gray-700">
              Page <span className="font-medium">{state.pageIndex + 1}</span> of{" "}
              <span className="font-medium">{pageOptions.length}</span>
            </span>
            <label>
              <span className="sr-only">Items Per Page</span>
              <select
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-300 focus:ring focus:ring-indigo-200 focus:ring-opacity-50"
                value={state.pageSize}
                onChange={(e) => {
                  setPageSize(Number(e.target.value));
                }}
              >
                {[5, 10, 15, 20].map((pageSize) => (
                  <option key={pageSize} value={pageSize}>
                    Show {pageSize}
                  </option>
                ))}
              </select>
            </label>
          </div>
          <div>
            <nav
              className="relative z-0 inline-flex rounded-md shadow-sm -space-x-px"
              aria-label="Pagination"
            >
              <PageButton
                className="rounded-l-md"
                onClick={() => gotoPage(0)}
                disabled={!canPreviousPage}
              >
                <span className="sr-only">First</span>
                <ChevronDoubleLeftIcon
                  className="h-5 w-5 text-gray-400"
                  aria-hidden="true"
                />
              </PageButton>
              <PageButton
                onClick={() => previousPage()}
                disabled={!canPreviousPage}
              >
                <span className="sr-only">Previous</span>
                <ChevronLeftIcon
                  className="h-5 w-5 text-gray-400"
                  aria-hidden="true"
                />
              </PageButton>
              <PageButton onClick={() => nextPage()} disabled={!canNextPage}>
                <span className="sr-only">Next</span>
                <ChevronRightIcon
                  className="h-5 w-5 text-gray-400"
                  aria-hidden="true"
                />
              </PageButton>
              <PageButton
                className="rounded-r-md"
                onClick={() => gotoPage(pageCount - 1)}
                disabled={!canNextPage}
              >
                <span className="sr-only">Last</span>
                <ChevronDoubleRightIcon
                  className="h-5 w-5 text-gray-400"
                  aria-hidden="true"
                />
              </PageButton>
            </nav>
          </div>
        </div>
      </div>
    </>
  );
}

export default Table;
