import React from "react";

const baseClass = "upload-list";

interface IUploadListProps {
  listItems: any[]; // TODO: typings
  HeadingComponent?: (props: any) => JSX.Element; // TODO: Typings
  ListItemComponent: (props: { listItem: any }) => JSX.Element; // TODO: types
  sortCompareFn?: (a: any, b: any) => number;
}

const UploadList = ({
  listItems,
  HeadingComponent,
  ListItemComponent,
  sortCompareFn,
}: IUploadListProps) => {
  const items = listItems.map((listItem) => {
    return (
      <li key={`${listItem.id}`} className={`${baseClass}__list-item`}>
        <ListItemComponent listItem={listItem} />
      </li>
    );
  });
  return (
    <div className={baseClass}>
      {HeadingComponent && (
        <div className={`${baseClass}__header`}>
          <HeadingComponent />
        </div>
      )}
      <ul className={`${baseClass}__list`}>
        {sortCompareFn ? items.sort(sortCompareFn) : items}
      </ul>
    </div>
  );
};

export default UploadList;
