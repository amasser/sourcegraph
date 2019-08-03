import { LocationDescriptor } from 'history'
import React from 'react'
import { LinkOrSpan } from '../../../../shared/src/components/LinkOrSpan'
import { pluralize } from '../../../../shared/src/util/strings'

export interface SummaryCountItemDescriptor<C> {
    noun: string
    pluralNoun?: string
    icon: React.ComponentType<{ className?: string }>
    count: number | ((context: C) => number)
    url?: LocationDescriptor | ((context: C) => LocationDescriptor)
    condition?: (context: C) => boolean
}

interface Props<C> {
    itemDescriptors: SummaryCountItemDescriptor<C>[]
    context: C

    className?: string
}

/**
 * A horizontal bar with item counts.
 */
export const SummaryCountBar = <C extends {}>({
    itemDescriptors,
    context,
    className = '',
}: Props<C>): React.ReactElement => (
    <nav className={`summary-count-bar border ${className}`}>
        <ul className="nav w-100">
            {itemDescriptors
                .filter(({ condition }) => !condition || condition(context))
                .map(({ icon: Icon, ...item }, i) => {
                    const count = typeof item.count === 'function' ? item.count(context) : item.count
                    return (
                        <li key={i} className="nav-item flex-1 text-center">
                            <LinkOrSpan
                                to={typeof item.url === 'function' ? item.url(context) : item.url}
                                className="nav-link"
                            >
                                <Icon className="icon-inline text-muted" /> <strong>{count}</strong>{' '}
                                <span className="text-muted">{pluralize(item.noun, count || 0, item.pluralNoun)}</span>
                            </LinkOrSpan>
                        </li>
                    )
                })}
        </ul>
    </nav>
)
