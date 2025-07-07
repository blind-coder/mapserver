package sqlite

const getBlocksByMtimeQuery = `
select x,y,z,data,mtime
from blocks b
where b.mtime > ?
order by b.mtime asc
limit ?
`

const countBlocksQuery = `
select count(*) from blocks b
`

const getBlockQuery = `
select x,y,z,data,mtime from blocks b where b.pos = ?
`

const getTimestampQuery = `
select strftime('%s', 'now')
`
