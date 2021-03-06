import * as GQL from '../../../../shared/src/graphql/schema'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import CheckIcon from 'mdi-react/CheckIcon'
import ClockOutlineIcon from 'mdi-react/ClockOutlineIcon'
import React, { FunctionComponent, useEffect, useMemo, useState } from 'react'
import { asError, ErrorLike, isErrorLike } from '../../../../shared/src/util/errors'
import { catchError, takeWhile, concatMap, repeatWhen, delay } from 'rxjs/operators'
import { ErrorAlert } from '../../components/alerts'
import { eventLogger } from '../../tracking/eventLogger'
import { fetchLsifIndex, deleteLsifIndex } from './backend'
import { Link } from '../../../../shared/src/components/Link'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { PageTitle } from '../../components/PageTitle'
import { RouteComponentProps, Redirect } from 'react-router'
import { Timestamp } from '../../components/time/Timestamp'
import { useObservable } from '../../../../shared/src/util/useObservable'
import DeleteIcon from 'mdi-react/DeleteIcon'
import { SchedulerLike, timer } from 'rxjs'
import * as H from 'history'

const REFRESH_INTERVAL_MS = 5000

interface Props extends RouteComponentProps<{ id: string }> {
    repo?: GQL.IRepository

    /** Scheduler for the refresh timer */
    scheduler?: SchedulerLike
    history: H.History
}

const terminalStates = new Set([GQL.LSIFIndexState.COMPLETED, GQL.LSIFIndexState.ERRORED])

function shouldReload(index: GQL.ILSIFIndex | ErrorLike | null | undefined): boolean {
    return !isErrorLike(index) && !(index && terminalStates.has(index.state))
}

/**
 * A page displaying metadata about an LSIF index.
 */
export const CodeIntelIndexPage: FunctionComponent<Props> = ({
    repo,
    scheduler,
    match: {
        params: { id },
    },
    history,
}) => {
    useEffect(() => eventLogger.logViewEvent('CodeIntelIndex'))

    const [deletionOrError, setDeletionOrError] = useState<'loading' | 'deleted' | ErrorLike>()

    const indexOrError = useObservable(
        useMemo(
            () =>
                timer(0, REFRESH_INTERVAL_MS, scheduler).pipe(
                    concatMap(() =>
                        fetchLsifIndex({ id }).pipe(
                            catchError((error): [ErrorLike] => [asError(error)]),
                            repeatWhen(observable => observable.pipe(delay(REFRESH_INTERVAL_MS)))
                        )
                    ),
                    takeWhile(shouldReload, true)
                ),
            [id, scheduler]
        )
    )

    const deleteIndex = async (): Promise<void> => {
        if (!indexOrError || isErrorLike(indexOrError)) {
            return
        }

        const description = `commit ${indexOrError.inputCommit.slice(0, 7)}`

        if (!window.confirm(`Delete auto-index record for commit ${description}?`)) {
            return
        }

        setDeletionOrError('loading')

        try {
            await deleteLsifIndex({ id }).toPromise()
            setDeletionOrError('deleted')
        } catch (error) {
            setDeletionOrError(error)
        }
    }

    return deletionOrError === 'deleted' ? (
        <Redirect to=".." />
    ) : isErrorLike(deletionOrError) ? (
        <ErrorAlert prefix="Error deleting LSIF index record" error={deletionOrError} history={history} />
    ) : (
        <div className="site-admin-lsif-index-page w-100">
            <PageTitle title="Code intelligence - auto-indexing" />
            {isErrorLike(indexOrError) ? (
                <ErrorAlert prefix="Error loading LSIF index" error={indexOrError} history={history} />
            ) : !indexOrError ? (
                <LoadingSpinner className="icon-inline" />
            ) : (
                <>
                    <div className="mb-1">
                        <h2 className="mb-0">
                            Auto-index record for commit{' '}
                            {indexOrError.projectRoot
                                ? indexOrError.projectRoot.commit.abbreviatedOID
                                : indexOrError.inputCommit.slice(0, 7)}{' '}
                            rooted at {indexOrError.projectRoot?.path || '/'}
                        </h2>
                    </div>

                    {indexOrError.state === GQL.LSIFIndexState.PROCESSING ? (
                        <div className="alert alert-primary mb-4 mt-3">
                            <LoadingSpinner className="icon-inline" />{' '}
                            <span className="e2e-index-state">Index is currently being processed...</span>
                        </div>
                    ) : indexOrError.state === GQL.LSIFIndexState.COMPLETED ? (
                        <div className="alert alert-success mb-4 mt-3">
                            <CheckIcon className="icon-inline" />{' '}
                            <span className="e2e-index-state">Index processed successfully.</span>
                        </div>
                    ) : indexOrError.state === GQL.LSIFIndexState.ERRORED ? (
                        <div className="alert alert-danger mb-4 mt-3">
                            <AlertCircleIcon className="icon-inline" />{' '}
                            <span className="e2e-index-state">Index failed to complete:</span>{' '}
                            <code>{indexOrError.failure}</code>
                        </div>
                    ) : (
                        <div className="alert alert-primary mb-4 mt-3">
                            <ClockOutlineIcon className="icon-inline" />{' '}
                            <span className="e2e-index-state">
                                Index is queued. There are {indexOrError.placeInQueue} indexes ahead of this one.
                            </span>
                        </div>
                    )}

                    <table className="table">
                        <tbody>
                            <tr>
                                <td>Repository</td>
                                <td>
                                    {indexOrError.projectRoot ? (
                                        <Link to={indexOrError.projectRoot.commit.repository.url}>
                                            {indexOrError.projectRoot.commit.repository.name}
                                        </Link>
                                    ) : (
                                        repo?.name || 'unknown'
                                    )}
                                </td>
                            </tr>

                            <tr>
                                <td>Commit</td>
                                <td>
                                    {indexOrError.projectRoot ? (
                                        <Link to={indexOrError.projectRoot.commit.url}>
                                            <code>{indexOrError.projectRoot.commit.oid}</code>
                                        </Link>
                                    ) : (
                                        indexOrError.inputCommit
                                    )}
                                </td>
                            </tr>

                            <tr>
                                <td>Queued</td>
                                <td>
                                    <Timestamp date={indexOrError.queuedAt} noAbout={true} />
                                </td>
                            </tr>

                            <tr>
                                <td>Began processing</td>
                                <td>
                                    {indexOrError.startedAt ? (
                                        <Timestamp date={indexOrError.startedAt} noAbout={true} />
                                    ) : (
                                        <span className="text-muted">Index has not yet started.</span>
                                    )}
                                </td>
                            </tr>

                            <tr>
                                <td>
                                    {indexOrError.state === GQL.LSIFIndexState.ERRORED && indexOrError.finishedAt
                                        ? 'Failed'
                                        : 'Finished'}{' '}
                                    processing
                                </td>
                                <td>
                                    {indexOrError.finishedAt ? (
                                        <Timestamp date={indexOrError.finishedAt} noAbout={true} />
                                    ) : (
                                        <span className="text-muted">Index has not yet completed.</span>
                                    )}
                                </td>
                            </tr>
                        </tbody>
                    </table>

                    <div className="action-container">
                        <div className="action-container__row">
                            <div className="action-container__description">
                                <h4 className="action-container__title">Delete this index</h4>
                                <div>Deleting this index will remove it from the index queue.</div>
                            </div>
                            <div className="action-container__btn-container">
                                <button
                                    type="button"
                                    className="btn btn-danger action-container__btn"
                                    onClick={deleteIndex}
                                    disabled={deletionOrError === 'loading'}
                                    data-tooltip="Delete index"
                                >
                                    <DeleteIcon className="icon-inline" />
                                </button>
                            </div>
                        </div>
                    </div>
                </>
            )}
        </div>
    )
}
