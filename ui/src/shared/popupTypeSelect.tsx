import { Autocomplete, EmptyState, ListBox, Popover, SearchField } from "@heroui/react";
import { useState, type CSSProperties } from "react";
import { exactMatch, useSearchFilter } from "./filter";

export type PopupTypeSelectProps<T> = {
    children: React.ReactNode;
    style?: CSSProperties;
    onExpanded?: (expanded: boolean) => void;
    onSelect?: (v: PopupTypeSelectOption<T>) => void;
    allowCreate?: boolean;
    value?: string;
    options: PopupTypeSelectOption<T>[]; // Displayed in order
    placement?: 'top' | 'left' | 'right' | 'bottom';
    maxMatches?: number; // Stops searching once this many options match
};

export type PopupTypeSelectOption<T> = {
    key: string;
    value?: T;
};

const DEFAULT_MAX_MATCHES = 50;

export function PopupTypeSelect<T>(props: PopupTypeSelectProps<T>) {
    const [isOpen, setIsOpen] = useState<boolean>(false);
    const [search, setSearch] = useState<string>('');

    const maxMatches = props.maxMatches ?? DEFAULT_MAX_MATCHES;

    // Stop scanning at maxMatches so a long options list costs no more than a short one.
    const matches = useSearchFilter(search, props.options, maxMatches);

    function setExpanded(expanded: boolean) {
        setIsOpen(expanded);
        if (!expanded) setSearch('');
        props.onExpanded?.(expanded);
    }

    function hasExactMatch(v: string): boolean {
        for (const match of matches) {
            if (exactMatch(v, match.key)) {
                return true;
            }
        }

        return false;
    }

    const listBoxStyle: CSSProperties = {
        maxHeight: 300,
        overflow: 'scroll',
    };

    return (
        <Popover isOpen={isOpen} onOpenChange={setExpanded}>
            <Popover.Trigger>
                {props.children}
            </Popover.Trigger>

            <Popover.Content placement={props.placement || 'bottom'}>
                <Popover.Arrow/>
                <Autocomplete.Filter inputValue={search} onInputChange={setSearch}>
                    <SearchField autoFocus name="search" variant="secondary">
                    <SearchField.Group>
                        <SearchField.SearchIcon />
                        <SearchField.Input placeholder="Search..." />
                        <SearchField.ClearButton />
                    </SearchField.Group>
                    </SearchField>
                    <ListBox
                        selectionMode="single"
                        disallowEmptySelection
                        style={listBoxStyle}
                        selectedKeys={props.value ? [props.value] : undefined}
                        onSelectionChange={(keys) => {
                            if (keys === 'all') return;

                            let selected = props.options.find((o) => keys.has(o.key));
                            if (keys.has(-1)) {
                                selected = {key: search};
                            }
                            if (selected) props.onSelect?.(selected);
                            setExpanded(false);
                        }}
                        renderEmptyState={() => <EmptyState>No results found</EmptyState>}
                    >
                        {props.allowCreate && search !== '' && !hasExactMatch(search) && <ListBox.Item id={-1} textValue={""}>
                            Create: {search}
                            <ListBox.ItemIndicator />
                        </ListBox.Item>}
                        {matches.map((item) => (
                            <ListBox.Item key={item.key} id={item.key} textValue={item.key}>
                                {item.key}
                                <ListBox.ItemIndicator />
                            </ListBox.Item>
                        ))}
                    </ListBox>
                </Autocomplete.Filter>
            </Popover.Content>
        </Popover>
    );
}

