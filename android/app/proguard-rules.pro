# R8 full mode removes the generated no-arg constructors that Room creates
# reflectively. WorkManager (pulled in by the Google Mobile Ads SDK) then fails
# in androidx.startup with "Failed to create an instance of WorkDatabase" and
# the release app crashes at launch.
-keep class * extends androidx.room.RoomDatabase { <init>(); }
-keep class androidx.work.impl.WorkDatabase_Impl { <init>(); }
-dontwarn androidx.room.paging.**
