import { nameToEmoji } from "gemoji";

export function classNames(...classes) {
  return classes.filter(Boolean).join(" ");
}

export function growthCalc(data) {
  if (data.baseline === 0) return 0;
  return (data.baseline + data.added) / data.baseline;
}

export function humanNumber(number) {
  if (number <= 1000) return `${number}`;
  let result, unit;
  if (number > 1000000) {
    result = `${(number / 1000000).toFixed(1)}`;
    unit = "M";
  } else if (number > 1000) {
    result = `${(number / 1000).toFixed(0)}`;
    unit = "k";
  }
  return parseFloat(result) + unit;
}

export function compareBasic(a, b) {
  return a === b ? 0 : a > b ? 1 : -1;
}

export function dataCompare(rowA, rowB, columnId) {
  let dataA = rowA.values[columnId].added;
  let dataB = rowB.values[columnId].added;
  return compareBasic(dataA, dataB);
}

export function sumNumberObjects(leafValues) {
  return leafValues.reduce((prev, cur) => {
    let newo = {};
    for (let key in prev) newo[key] = prev[key] + cur[key];
    return newo;
  });
}

export function massageData(data) {
  // Unfortunately have to combine added / baseline data into 1 value in
  // order to be able to groupby and recalculate growth rates
  return data.map((obj) => {
    if (obj.Description.length > 0 && !containsChinese(obj.Description))
      return {
        Name: obj.Name,
        Stars: obj.Stars,
        Language: obj.Language,
        Description: obj.Description,
        data: {
          Added7: obj.Added7,
          Baseline7: obj.Baseline7,
          Added30: obj.Added30,
          Baseline30: obj.Baseline30,
          Added90: obj.Added90,
          Baseline90: obj.Baseline90,
        },
      };
  }).filter(Boolean);
}

export function equalsForSelect(rows, ids, filterValue) {
  return rows.filter((row) => {
    return ids.some((id) => {
      const rowValue = row.values[id];
      // eslint-disable-next-line eqeqeq
      return rowValue !== "" && (rowValue == filterValue || filterValue === "");
    });
  });
}

equalsForSelect.autoRemove = (val) => val == null;

export function replaceColonmoji(text) {
  return text.replace(/:[^:\s]*(?:::[^:\s]*)*:/g, function (match, capture) {
    return nameToEmoji[match.slice(1, -1)] || match;
  });
}

export function containsChinese(text) {
  return /[\u4E00-\u9FCC\u3400-\u4DB5\uFA0E\uFA0F\uFA11\uFA13\uFA14\uFA1F\uFA21\uFA23\uFA24\uFA27-\uFA29]|[\ud840-\ud868][\udc00-\udfff]|\ud869[\udc00-\uded6\udf00-\udfff]|[\ud86a-\ud86c][\udc00-\udfff]|\ud86d[\udc00-\udf34\udf40-\udfff]|\ud86e[\udc00-\udc1d]/.test(
    text
  );
}
