import { IngredientUnit, RecipeTimeUnit, type RecipeTime } from "../apiTypes";

export function timeUnitLabel(unit: RecipeTimeUnit): string {
    let unitLabel = "sec";
    if (unit === RecipeTimeUnit.Minutes) {
        unitLabel = "min";
    } else if (unit === RecipeTimeUnit.Hours) {
        unitLabel = "hr";
    } else if (unit === RecipeTimeUnit.Days) {
        unitLabel = "days";
    } else if (unit === RecipeTimeUnit.Weeks) {
        unitLabel = "wks";
    }
    return unitLabel;
}

export function fullTimeLabel(time?: RecipeTime): string {
    if (!time) {
        return "??time??";
    }

    const singular = time.start_time === 1 && !time.end_time;

    let unitLabel = "second";
    if (time.unit === RecipeTimeUnit.Minutes) {
        unitLabel = "minute";
    } else if (time.unit === RecipeTimeUnit.Hours) {
        unitLabel = "hour";
    } else if (time.unit === RecipeTimeUnit.Days) {
        unitLabel = "day";
    } else if (time.unit === RecipeTimeUnit.Weeks) {
        unitLabel = "week";
    }

    if (!singular) {
        unitLabel = unitLabel + 's';
    }

    if (time.end_time) {
        return time.start_time.toString() + '-' + time.end_time.toString() + ' ' + unitLabel;
    }

    return time.start_time.toString() + ' ' + unitLabel;
}

export function quantityLabel(value: number, unit?: IngredientUnit): string {
    const unitLabel: unitDetails = quantityUnitLabel(unit ?? IngredientUnit.Unspecified);
    let valueStr = value.toString();
    let fracStr = '';
    const wholeValue = Math.floor(value);
    const fracValue = value - wholeValue;
    if (Math.abs(fracValue - 0.333) < 0.01) {
        fracStr = '⅓';
    } else if (Math.abs(fracValue - 0.25) < 0.01) {
        fracStr = '¼';
    } else if (Math.abs(fracValue - 0.5) < 0.01) {
        fracStr = '½';
    } else if (Math.abs(fracValue - 0.666) < 0.01) {
        fracStr = '⅔';
    } else if (Math.abs(fracValue - 0.75) < 0.01) {
        fracStr = '¾';
    }

    if (fracStr !== '' && wholeValue > 0) {
        valueStr = wholeValue.toString() + fracStr;
    } else if (fracStr !== '') {
        valueStr = fracStr;
    }

    if (unitLabel.singular === '') {
        return valueStr;
    } else if (value === 1) {
        return valueStr + ' ' + unitLabel.singular;
    } else {
        return valueStr + ' ' + unitLabel.plural;
    }
}

export type unitDetails = {
    id: IngredientUnit;
    singular: string;
    plural: string;
    metric: boolean;
    volume: boolean;
    usesQuantity: boolean;
};

export const validUnits: Record<IngredientUnit, unitDetails> = {
    [IngredientUnit.Grams]: {
        id: IngredientUnit.Grams,
        singular: 'gram',
        plural: 'grams',
        metric: true,
        volume: false,
        usesQuantity: true,
    },
    [IngredientUnit.Kilograms]: {
        id: IngredientUnit.Kilograms,
        singular: 'kg',
        plural: 'kg',
        metric: true,
        volume: false,
        usesQuantity: true,
    },
    [IngredientUnit.Milligrams]: {
        id: IngredientUnit.Milligrams,
        singular: 'mg',
        plural: 'mg',
        metric: true,
        volume: false,
        usesQuantity: true,
    },
    [IngredientUnit.Ounces]: {
        id: IngredientUnit.Ounces,
        singular: 'oz',
        plural: 'oz',
        metric: false,
        volume: false,
        usesQuantity: true,
    },
    [IngredientUnit.Pounds]: {
        id: IngredientUnit.Pounds,
        singular: 'lb',
        plural: 'lbs',
        metric: false,
        volume: false,
        usesQuantity: true,
    },
    [IngredientUnit.Milliliters]: {
        id: IngredientUnit.Milliliters,
        singular: 'ml',
        plural: 'ml',
        metric: true,
        volume: true,
        usesQuantity: true,
    },
    [IngredientUnit.Liters]: {
        id: IngredientUnit.Liters,
        singular: 'liter',
        plural: 'liters',
        metric: true,
        volume: true,
        usesQuantity: true,
    },
    [IngredientUnit.Cups]: {
        id: IngredientUnit.Cups,
        singular: 'cup',
        plural: 'cups',
        metric: false,
        volume: true,
        usesQuantity: true,
    },
    [IngredientUnit.Tablespoons]: {
        id: IngredientUnit.Tablespoons,
        singular: 'tbsp',
        plural: 'tbsp',
        metric: false,
        volume: true,
        usesQuantity: true,
    },
    [IngredientUnit.Teaspoons]: {
        id: IngredientUnit.Teaspoons,
        singular: 'tsp',
        plural: 'tsp',
        metric: false,
        volume: true,
        usesQuantity: true,
    },
    [IngredientUnit.FluidOunces]: {
        id: IngredientUnit.FluidOunces,
        singular: 'fl oz',
        plural: 'fl oz',
        metric: false,
        volume: true,
        usesQuantity: true,
    },
    [IngredientUnit.Gallons]: {
        id: IngredientUnit.Gallons,
        singular: 'gal',
        plural: 'gal',
        metric: false,
        volume: true,
        usesQuantity: true,
    },
    [IngredientUnit.Quarts]: {
        id: IngredientUnit.Quarts,
        singular: 'quart',
        plural: 'quarts',
        metric: false,
        volume: true,
        usesQuantity: true,
    },
    [IngredientUnit.Pieces]: {
        id: IngredientUnit.Pieces,
        singular: '',
        plural: '',
        metric: false,
        volume: false,
        usesQuantity: true,
    },
    [IngredientUnit.Pinches]: {
        id: IngredientUnit.Pinches,
        singular: 'pinch',
        plural: 'pinches',
        metric: false,
        volume: true,
        usesQuantity: true,
    },
    [IngredientUnit.Containers]: {
        id: IngredientUnit.Containers,
        singular: 'container',
        plural: 'containers',
        metric: false,
        volume: false,
        usesQuantity: true,
    },
    [IngredientUnit.Unspecified]: {
        id: IngredientUnit.Unspecified,
        singular: '',
        plural: '',
        metric: false,
        volume: false,
        usesQuantity: false,
    },
    [IngredientUnit.Unknown]: {
        id: IngredientUnit.Unknown,
        singular: '',
        plural: '',
        metric: false,
        volume: false,
        usesQuantity: false,
    },
};

export const unknownUnit: unitDetails = {
    id: IngredientUnit.Unknown,
    singular: '??',
    plural: '??',
    metric: false,
    volume: false,
    usesQuantity: false,
};

export function quantityUnitLabel(unit: IngredientUnit): unitDetails {
    return validUnits[unit] ?? unknownUnit;
}