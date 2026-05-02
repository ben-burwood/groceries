export interface Grocery {
    uuid: string;
    name: string;
}

export type PendingCreate = Grocery;

export interface PendingNeeded {
    uuid: string;
    needed: boolean;
}

export type GroceryItem = Grocery & {
    needed: boolean;
    pending?: boolean;
};
