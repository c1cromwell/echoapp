package com.echo.shared_data.state

import com.echo.shared_data.types.MerkleRootStatus

import java.util.concurrent.ConcurrentHashMap
import scala.jdk.CollectionConverters._

/** In-memory index for WO-230 Step 5 finality polling on Data L1. */
object DataMerkleRootIndex {

  private val roots = new ConcurrentHashMap[String, MerkleRootStatus]()

  def record(root: String, leafCount: Int, finalized: Boolean): Unit = {
    // Never downgrade: a root that L0 has folded stays finalized even if the
    // same root is re-submitted and re-validated afterwards.
    roots.merge(
      root,
      MerkleRootStatus(root = root, leafCount = leafCount, finalized = finalized),
      (old, neu) => if (old.finalized) old else neu
    )
    ()
  }

  def lookup(root: String): Option[MerkleRootStatus] =
    Option(roots.get(root))

  def clear(): Unit =
    roots.clear()

  def snapshot: Map[String, MerkleRootStatus] =
    roots.asScala.toMap
}
