#!/usr/bin/env perl

use warnings;
use strict;
use 5.10.0;

use Getopt::Long;

my $project_name = 'project';
my $distribution = 'stable';
my $editor = 'DigitalOcean CI <systat@digitalocean.com>';

GetOptions(
  "project:s"      => \$project_name,
  "distribution:s" => \$distribution,
  "editor:s"       => \$editor,
);

my @tags = split("\n", `git tag --sort version:refname --merged`);

my @hunks;
my $previous_tag = "";

foreach my $tag (@tags) {
	my $hunk;

	# Go needs the v in the tag name. Debian hates it. This lets us support both
	my $deb_tag = $tag;
	$deb_tag = (split /\//, $tag)[-1];
	$deb_tag =~ s/^v//;

	$hunk .= "$project_name ($deb_tag) $distribution; urgency=medium\n";

	my @refs;
	if ($previous_tag eq "") {
		@refs = split("\n", `git log --pretty=format:%h --no-merges $tag`);
	} else {
		@refs = split("\n", `git log --pretty=format:%h --no-merges $previous_tag..$tag`);
	}
	$previous_tag = $tag;


	my $prev_author = "";
	for my $ref (reverse @refs) {
		my $author = `git show -s --pretty=format:%an $ref`;
		if ($prev_author ne $author) {
			$hunk .= "\n  [ $author ]\n";
		}
		$prev_author = $author;

		$hunk .= format_ref($ref)
	}

	my $date = `git show -s --pretty=format:%cD "$tag^{commit}"`;
	$hunk .= "\n -- $editor  $date\n\n";

	push @hunks, $hunk;
}

my @refs;
my $most_recent_tag = $tags[-1];
if (scalar @tags > 0) {
	@refs = split("\n", `git log --pretty=format:%h --no-merges $most_recent_tag..HEAD`);
} else {
	@refs = split("\n", `git log --pretty=format:%h --no-merges`);
	$most_recent_tag = "0.0.0";
}

my $len = scalar(@refs);
if ($len > 0) {
	$most_recent_tag = (split /\//, $most_recent_tag)[-1];
	$most_recent_tag =~ s/^v//;

	my $hunk = "$project_name ($most_recent_tag+$len~$refs[0]) unstable; urgency=medium\n";
	my $prev_author = "";

	for my $ref (reverse @refs) {
		my $author = `git show -s --pretty=format:%an $ref`;
		if ($prev_author ne $author) {
			$hunk .= "\n  [ $author ]\n";
		}
		$prev_author = $author;

		$hunk .= format_ref($ref);
	}

	my $date = `git show -s --pretty=format:%cD "$refs[0]"`;
	$hunk .= "\n -- $editor  $date\n\n";
	push @hunks, $hunk;
}

for my $hunk (reverse @hunks) {
	print $hunk;
}

###########

sub format_ref {
	my $ref = shift or return "";
	my $hunk = "  * ";
	my $count = 0;

	for my $line (split("\n", `git show -s --pretty=format:'%s' $ref | fold -s -w 72`)) {
		chomp $line;
		if (!$count) {
			$hunk .= "$line\n";
		} else {
			$hunk .= "    $line\n";
		}
		$count++;
	}

	return $hunk;
}
